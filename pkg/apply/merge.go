package apply

import (
	"github.com/pkg/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func MergeObjectForUpdate(current, updated *uns.Unstructured) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	updated.SetResourceVersion(current.GetResourceVersion())
	if err := MergeDeploymentForUpdate(current, updated); err != nil {
		return err
	}
	if err := MergeServiceForUpdate(current, updated); err != nil {
		return err
	}
	if err := MergeServiceAccountForUpdate(current, updated); err != nil {
		return err
	}
	mergeAnnotations(current, updated)
	mergeLabels(current, updated)
	return nil
}

const (
	deploymentRevisionAnnotation = "deployment.kubernetes.io/revision"
)

func MergeDeploymentForUpdate(current, updated *uns.Unstructured) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	gvk := updated.GroupVersionKind()
	if gvk.Group == "apps" && gvk.Kind == "Deployment" {
		curAnnotations := current.GetAnnotations()
		updatedAnnotations := updated.GetAnnotations()
		if updatedAnnotations == nil {
			updatedAnnotations = map[string]string{}
		}
		anno, ok := curAnnotations[deploymentRevisionAnnotation]
		if ok {
			updatedAnnotations[deploymentRevisionAnnotation] = anno
		}
		updated.SetAnnotations(updatedAnnotations)
	}
	return nil
}
func MergeServiceForUpdate(current, updated *uns.Unstructured) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	gvk := updated.GroupVersionKind()
	if gvk.Group == "" && gvk.Kind == "Service" {
		clusterIP, found, err := uns.NestedString(current.Object, "spec", "clusterIP")
		if err != nil {
			return err
		}
		if found {
			return uns.SetNestedField(updated.Object, clusterIP, "spec", "clusterIP")
		}
	}
	return nil
}
func MergeServiceAccountForUpdate(current, updated *uns.Unstructured) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	gvk := updated.GroupVersionKind()
	if gvk.Group == "" && gvk.Kind == "ServiceAccount" {
		curSecrets, ok, err := uns.NestedSlice(current.Object, "secrets")
		if err != nil {
			return err
		}
		if ok {
			uns.SetNestedField(updated.Object, curSecrets, "secrets")
		}
	}
	return nil
}
func mergeAnnotations(current, updated *uns.Unstructured) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	updatedAnnotations := updated.GetAnnotations()
	curAnnotations := current.GetAnnotations()
	if curAnnotations == nil {
		curAnnotations = map[string]string{}
	}
	for k, v := range updatedAnnotations {
		curAnnotations[k] = v
	}
	updated.SetAnnotations(curAnnotations)
}
func mergeLabels(current, updated *uns.Unstructured) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	updatedLabels := updated.GetLabels()
	curLabels := current.GetLabels()
	if curLabels == nil {
		curLabels = map[string]string{}
	}
	for k, v := range updatedLabels {
		curLabels[k] = v
	}
	updated.SetLabels(curLabels)
}
func IsObjectSupported(obj *uns.Unstructured) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	gvk := obj.GroupVersionKind()
	if gvk.Group == "" && gvk.Kind == "ServiceAccount" {
		secrets, ok, err := uns.NestedSlice(obj.Object, "secrets")
		if err != nil {
			return err
		}
		if ok && len(secrets) > 0 {
			return errors.Errorf("cannot create ServiceAccount with secrets")
		}
	}
	return nil
}
