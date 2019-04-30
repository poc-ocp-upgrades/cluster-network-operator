package k8s

import (
	"encoding/json"
	"github.com/pkg/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func ToUnstructured(obj interface{}) (*uns.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	b, err := json.Marshal(obj)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to convert to unstructured (marshal)")
	}
	u := &uns.Unstructured{}
	if err := json.Unmarshal(b, u); err != nil {
		return nil, errors.Wrapf(err, "failed to convert to unstructured (unmarshal)")
	}
	return u, nil
}
