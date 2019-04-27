package network

import (
	"fmt"
	"github.com/onsi/gomega/types"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func HaveKubernetesID(kind, namespace, name string) types.GomegaMatcher {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &KubeObjectMatcher{kind: kind, namespace: namespace, name: name}
}

type KubeObjectMatcher struct{ kind, namespace, name string }

func (k *KubeObjectMatcher) Match(actual interface{}) (bool, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, ok := actual.(*uns.Unstructured)
	if !ok {
		return false, fmt.Errorf("cannot match object of type %t", actual)
	}
	ok = obj.GetKind() == k.kind && obj.GetNamespace() == k.namespace && obj.GetName() == k.name
	return ok, nil
}
func (k *KubeObjectMatcher) FailureMessage(actual interface{}) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, ok := actual.(*uns.Unstructured)
	if !ok {
		return "not of type Unstructured"
	}
	return fmt.Sprintf("Expected Kind, Namespace, Name to match (%v, %v, %v) but got (%v, %v, %v)", k.kind, k.namespace, k.name, obj.GetKind(), obj.GetNamespace(), obj.GetName())
}
func (k *KubeObjectMatcher) NegatedFailureMessage(actual interface{}) string {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	obj, ok := actual.(*uns.Unstructured)
	if !ok {
		return "not of type Unstructured"
	}
	return fmt.Sprintf("Expected Kind, Namespace, Name not to match (%v, %v, %v) but got (%v, %v, %v)", k.kind, k.namespace, k.name, obj.GetKind(), obj.GetNamespace(), obj.GetName())
}
func tweakMetaForCompare(obj *uns.Unstructured) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	if obj.GetLabels() == nil {
		obj.SetLabels(map[string]string{})
	}
	if obj.GetAnnotations() == nil {
		obj.SetAnnotations(map[string]string{})
	}
	obj.SetResourceVersion("")
}
