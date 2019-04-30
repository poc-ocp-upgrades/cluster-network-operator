package controller

import (
	"github.com/openshift/cluster-network-operator/pkg/controller/statusmanager"
	operatorversion "github.com/openshift/cluster-network-operator/version"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

var AddToManagerFuncs []func(manager.Manager, *statusmanager.StatusManager) error

func AddToManager(m manager.Manager) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	s := statusmanager.New(m.GetClient(), "network", operatorversion.Version)
	for _, f := range AddToManagerFuncs {
		if err := f(m, s); err != nil {
			return err
		}
	}
	return nil
}
