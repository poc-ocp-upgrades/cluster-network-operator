package operconfig

import (
	"log"
	"github.com/openshift/cluster-network-operator/pkg/controller/statusmanager"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func newPodReconciler(status *statusmanager.StatusManager) *ReconcilePods {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return &ReconcilePods{status: status}
}

var _ reconcile.Reconciler = &ReconcilePods{}

type ReconcilePods struct {
	status		*statusmanager.StatusManager
	resources	[]types.NamespacedName
}

func (r *ReconcilePods) SetResources(resources []types.NamespacedName) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	r.resources = resources
}
func (r *ReconcilePods) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	found := false
	for _, name := range r.resources {
		if name.Namespace == request.Namespace && name.Name == request.Name {
			found = true
			break
		}
	}
	if !found {
		return reconcile.Result{}, nil
	}
	log.Printf("Reconciling update to %s/%s\n", request.Namespace, request.Name)
	r.status.SetFromPods()
	return reconcile.Result{RequeueAfter: ResyncPeriod}, nil
}
