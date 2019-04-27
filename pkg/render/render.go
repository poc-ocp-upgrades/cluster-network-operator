package render

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"github.com/Masterminds/sprig"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

type RenderData struct {
	Funcs	template.FuncMap
	Data	map[string]interface{}
}

func MakeRenderData() RenderData {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return RenderData{Funcs: template.FuncMap{}, Data: map[string]interface{}{}}
}
func RenderDir(manifestDir string, d *RenderData) ([]*unstructured.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	out := []*unstructured.Unstructured{}
	if err := filepath.Walk(manifestDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !(strings.HasSuffix(path, ".yml") || strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".json")) {
			return nil
		}
		objs, err := RenderTemplate(path, d)
		if err != nil {
			return err
		}
		out = append(out, objs...)
		return nil
	}); err != nil {
		return nil, errors.Wrap(err, "error rendering manifests")
	}
	return out, nil
}
func RenderTemplate(path string, d *RenderData) ([]*unstructured.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	tmpl := template.New(path).Option("missingkey=error")
	if d.Funcs != nil {
		tmpl.Funcs(d.Funcs)
	}
	tmpl.Funcs(template.FuncMap{"getOr": getOr, "isSet": isSet})
	tmpl.Funcs(sprig.TxtFuncMap())
	source, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read manifest %s", path)
	}
	if _, err := tmpl.Parse(string(source)); err != nil {
		return nil, errors.Wrapf(err, "failed to parse manifest %s as template", path)
	}
	rendered := bytes.Buffer{}
	if err := tmpl.Execute(&rendered, d.Data); err != nil {
		return nil, errors.Wrapf(err, "failed to render manifest %s", path)
	}
	out := []*unstructured.Unstructured{}
	if len(strings.TrimSpace(rendered.String())) == 0 {
		return out, nil
	}
	decoder := yaml.NewYAMLOrJSONDecoder(&rendered, 4096)
	for {
		u := unstructured.Unstructured{}
		if err := decoder.Decode(&u); err != nil {
			if err == io.EOF {
				break
			}
			return nil, errors.Wrapf(err, "failed to unmarshal manifest %s", path)
		}
		out = append(out, &u)
	}
	return out, nil
}
