package operconfig

import (
	"context"
	"fmt"
	"log"
	"time"
	"github.com/pkg/errors"
	configv1 "github.com/openshift/api/config/v1"
	operv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-network-operator/pkg/apply"
	"github.com/openshift/cluster-network-operator/pkg/controller/statusmanager"
	"github.com/openshift/cluster-network-operator/pkg/names"
	"github.com/openshift/cluster-network-operator/pkg/network"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var ResyncPeriod = 5 * time.Minute
var ManifestPath = "./bindata"

func Add(mgr manager.Manager, status *statusmanager.StatusManager) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return add(mgr, newReconciler(mgr, status))
}
func newReconciler(mgr manager.Manager, status *statusmanager.StatusManager) *ReconcileOperConfig {
	_logClusterCodePath()
	defer _logClusterCodePath()
	configv1.Install(mgr.GetScheme())
	operv1.Install(mgr.GetScheme())
	return &ReconcileOperConfig{client: mgr.GetClient(), scheme: mgr.GetScheme(), status: status, podReconciler: newPodReconciler(status)}
}
func add(mgr manager.Manager, r *ReconcileOperConfig) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	c, err := controller.New("operconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &operv1.Network{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	c, err = controller.New("pod-controller", mgr, controller.Options{Reconciler: r.podReconciler})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.DaemonSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

var _ reconcile.Reconciler = &ReconcileOperConfig{}

type ReconcileOperConfig struct {
	client			client.Client
	scheme			*runtime.Scheme
	status			*statusmanager.StatusManager
	podReconciler	*ReconcilePods
}

func (r *ReconcileOperConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	log.Printf("Reconciling Network.operator.openshift.io %s\n", request.Name)
	if request.Name != names.OPERATOR_CONFIG {
		log.Printf("Ignoring Network.operator.openshift.io without default name")
		return reconcile.Result{}, nil
	}
	operConfig := &operv1.Network{TypeMeta: metav1.TypeMeta{APIVersion: operv1.GroupVersion.String(), Kind: "Network"}}
	err := r.client.Get(context.TODO(), request.NamespacedName, operConfig)
	if err != nil {
		if apierrors.IsNotFound(err) {
			r.status.SetDegraded(statusmanager.OperatorConfig, "NoOperatorConfig", fmt.Sprintf("Operator configuration %s was deleted", request.NamespacedName.String()))
			return reconcile.Result{}, nil
		}
		log.Printf("Unable to retrieve Network.operator.openshift.io object: %v", err)
		return reconcile.Result{}, err
	}
	if err := r.MergeClusterConfig(context.TODO(), operConfig); err != nil {
		log.Printf("Failed to merge the cluster configuration: %v", err)
		r.status.SetDegraded(statusmanager.OperatorConfig, "MergeClusterConfig", fmt.Sprintf("Internal error while merging cluster configuration and operator configuration: %v", err))
		return reconcile.Result{}, err
	}
	network.Canonicalize(&operConfig.Spec)
	if err := network.Validate(&operConfig.Spec); err != nil {
		log.Printf("Failed to validate Network.operator.openshift.io.Spec: %v", err)
		r.status.SetDegraded(statusmanager.OperatorConfig, "InvalidOperatorConfig", fmt.Sprintf("The operator configuration is invalid (%v). Use 'oc edit network.operator.openshift.io cluster' to fix.", err))
		return reconcile.Result{}, err
	}
	prev, err := GetAppliedConfiguration(context.TODO(), r.client, operConfig.ObjectMeta.Name)
	if err != nil {
		log.Printf("Failed to retrieve previously applied configuration: %v", err)
		return reconcile.Result{}, err
	}
	network.FillDefaults(&operConfig.Spec, prev)
	if prev != nil {
		err = network.IsChangeSafe(prev, &operConfig.Spec)
		if err != nil {
			log.Printf("Not applying unsafe change: %v", err)
			r.status.SetDegraded(statusmanager.OperatorConfig, "InvalidOperatorConfig", fmt.Sprintf("Not applying unsafe configuration change: %v. Use 'oc edit network.operator.openshift.io cluster' to undo the change.", err))
			return reconcile.Result{}, err
		}
	}
	objs, err := network.Render(&operConfig.Spec, ManifestPath)
	if err != nil {
		log.Printf("Failed to render: %v", err)
		r.status.SetDegraded(statusmanager.OperatorConfig, "RenderError", fmt.Sprintf("Internal error while rendering operator configuration: %v", err))
		return reconcile.Result{}, err
	}
	app, err := AppliedConfiguration(operConfig)
	if err != nil {
		log.Printf("Failed to render applied: %v", err)
		r.status.SetDegraded(statusmanager.OperatorConfig, "RenderError", fmt.Sprintf("Internal error while recording new operator configuration: %v", err))
		return reconcile.Result{}, err
	}
	objs = append([]*uns.Unstructured{app}, objs...)
	daemonSets := []types.NamespacedName{}
	deployments := []types.NamespacedName{}
	for _, obj := range objs {
		if obj.GetAPIVersion() == "apps/v1" && obj.GetKind() == "DaemonSet" {
			daemonSets = append(daemonSets, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()})
		} else if obj.GetAPIVersion() == "apps/v1" && obj.GetKind() == "Deployment" {
			deployments = append(deployments, types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()})
		}
	}
	r.status.SetDaemonSets(daemonSets)
	r.status.SetDeployments(deployments)
	allResources := []types.NamespacedName{}
	allResources = append(allResources, daemonSets...)
	allResources = append(allResources, deployments...)
	r.podReconciler.SetResources(allResources)
	for _, obj := range objs {
		if err := controllerutil.SetControllerReference(operConfig, obj, r.scheme); err != nil {
			err = errors.Wrapf(err, "could not set reference for (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			log.Println(err)
			r.status.SetDegraded(statusmanager.OperatorConfig, "InternalError", fmt.Sprintf("Internal error while updating operator configuration: %v", err))
			return reconcile.Result{}, err
		}
		if err := apply.ApplyObject(context.TODO(), r.client, obj); err != nil {
			err = errors.Wrapf(err, "could not apply (%s) %s/%s", obj.GroupVersionKind(), obj.GetNamespace(), obj.GetName())
			log.Println(err)
			anno := obj.GetAnnotations()
			if anno != nil {
				if _, ok := anno[names.IgnoreObjectErrorAnnotation]; ok {
					log.Println("Object has ignore-errors annotation set, continuing")
					continue
				}
			}
			r.status.SetDegraded(statusmanager.OperatorConfig, "ApplyOperatorConfig", fmt.Sprintf("Error while updating operator configuration: %v", err))
			return reconcile.Result{}, err
		}
	}
	status, err := r.ClusterNetworkStatus(context.TODO(), operConfig)
	if err != nil {
		log.Printf("Could not generate network status: %v", err)
		r.status.SetDegraded(statusmanager.OperatorConfig, "StatusError", fmt.Sprintf("Could not update cluster configuration status: %v", err))
		return reconcile.Result{}, err
	}
	if status != nil {
		if err := apply.ApplyObject(context.TODO(), r.client, status); err != nil {
			err = errors.Wrapf(err, "could not apply (%s) %s/%s", status.GroupVersionKind(), status.GetNamespace(), status.GetName())
			log.Println(err)
			r.status.SetDegraded(statusmanager.OperatorConfig, "StatusError", fmt.Sprintf("Could not update cluster configuration status: %v", err))
			return reconcile.Result{}, err
		}
	}
	r.status.SetNotDegraded(statusmanager.OperatorConfig)
	return reconcile.Result{RequeueAfter: ResyncPeriod}, nil
}
