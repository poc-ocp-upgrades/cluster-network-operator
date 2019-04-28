package apply

import (
	"bytes"
	"testing"
	. "github.com/onsi/gomega"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
)

func TestMergeNamespace(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
  labels:
    a: cur
    b: cur
  annotations:
    a: cur
    b: cur`)
	upd := UnstructuredFromYaml(t, `
apiVersion: v1
kind: Namespace
metadata:
  name: ns1
  labels:
    a: upd
    c: upd
  annotations:
    a: upd
    c: upd`)
	err := MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(upd.GetLabels()).To(Equal(map[string]string{"a": "upd", "b": "cur", "c": "upd"}))
	g.Expect(upd.GetAnnotations()).To(Equal(map[string]string{"a": "upd", "b": "cur", "c": "upd"}))
}
func TestMergeDeployment(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: cur
    b: cur
  annotations:
    deployment.kubernetes.io/revision: cur
    a: cur
    b: cur`)
	upd := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: upd
    c: upd
  annotations:
    deployment.kubernetes.io/revision: upd
    a: upd
    c: upd`)
	err := MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(upd.GetLabels()).To(Equal(map[string]string{"a": "upd", "b": "cur", "c": "upd"}))
	g.Expect(upd.GetAnnotations()).To(Equal(map[string]string{"a": "upd", "b": "cur", "c": "upd", "deployment.kubernetes.io/revision": "cur"}))
}
func TestMergeNilCur(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)
	upd := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: upd
    c: upd
  annotations:
    a: upd
    c: upd`)
	err := MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(upd.GetLabels()).To(Equal(map[string]string{"a": "upd", "c": "upd"}))
	g.Expect(upd.GetAnnotations()).To(Equal(map[string]string{"a": "upd", "c": "upd"}))
}
func TestMergeNilMeta(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)
	upd := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)
	err := MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(upd.GetLabels()).To(BeEmpty())
}
func TestMergeNilUpd(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1
  labels:
    a: cur
    b: cur
  annotations:
    a: cur
    b: cur`)
	upd := UnstructuredFromYaml(t, `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: d1`)
	err := MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(upd.GetLabels()).To(Equal(map[string]string{"a": "cur", "b": "cur"}))
	g.Expect(upd.GetAnnotations()).To(Equal(map[string]string{"a": "cur", "b": "cur"}))
}
func TestMergeService(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: v1
kind: Service
metadata:
  name: d1
spec:
  clusterIP: cur`)
	upd := UnstructuredFromYaml(t, `
apiVersion: v1
kind: Service
metadata:
  name: d1
spec:
  clusterIP: upd`)
	err := MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	ip, _, err := uns.NestedString(upd.Object, "spec", "clusterIP")
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(ip).To(Equal("cur"))
}
func TestMergeServiceAccount(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	cur := UnstructuredFromYaml(t, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: d1
  annotations:
    a: cur
secrets:
- foo`)
	upd := UnstructuredFromYaml(t, `
apiVersion: v1
kind: ServiceAccount
metadata:
  name: d1
  annotations:
    b: upd`)
	err := IsObjectSupported(cur)
	g.Expect(err).To(MatchError(ContainSubstring("cannot create ServiceAccount with secrets")))
	err = MergeObjectForUpdate(cur, upd)
	g.Expect(err).NotTo(HaveOccurred())
	s, ok, err := uns.NestedSlice(upd.Object, "secrets")
	g.Expect(ok).To(BeTrue())
	g.Expect(s).To(ConsistOf("foo"))
}
func UnstructuredFromYaml(t *testing.T, obj string) *uns.Unstructured {
	_logClusterCodePath()
	defer _logClusterCodePath()
	t.Helper()
	buf := bytes.NewBufferString(obj)
	decoder := yaml.NewYAMLOrJSONDecoder(buf, 4096)
	u := uns.Unstructured{}
	err := decoder.Decode(&u)
	if err != nil {
		t.Fatalf("failed to parse test yaml: %v", err)
	}
	return &u
}
