package network

import (
	"testing"
	operv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-network-operator/pkg/apply"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	. "github.com/onsi/gomega"
)

var OVNKubernetesConfig = operv1.Network{Spec: operv1.NetworkSpec{ServiceNetwork: []string{"172.30.0.0/16"}, ClusterNetwork: []operv1.ClusterNetworkEntry{{CIDR: "10.128.0.0/15", HostPrefix: 23}, {CIDR: "10.0.0.0/14", HostPrefix: 24}}, DefaultNetwork: operv1.DefaultNetworkDefinition{Type: operv1.NetworkTypeOVNKubernetes, OVNKubernetesConfig: &operv1.OVNKubernetesConfig{}}}}
var manifestDirOvn = "../../bindata"

func TestRenderOVNKubernetes(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	crd := OVNKubernetesConfig.DeepCopy()
	config := &crd.Spec
	errs := validateOVNKubernetes(config)
	g.Expect(errs).To(HaveLen(0))
	FillDefaults(config, nil)
	objs, err := renderOVNKubernetes(config, manifestDirOvn)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(objs).To(ContainElement(HaveKubernetesID("DaemonSet", "openshift-ovn-kubernetes", "ovnkube-node")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("Deployment", "openshift-ovn-kubernetes", "ovnkube-master")))
	g.Expect(objs[0]).To(HaveKubernetesID("Namespace", "", "openshift-ovn-kubernetes"))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("ClusterRole", "", "openshift-ovn-kubernetes-node")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("ClusterRole", "", "openshift-ovn-kubernetes-controller")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("ServiceAccount", "openshift-ovn-kubernetes", "ovn-kubernetes-node")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("ServiceAccount", "openshift-ovn-kubernetes", "ovn-kubernetes-controller")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("ClusterRoleBinding", "", "openshift-ovn-kubernetes-node")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("Deployment", "openshift-ovn-kubernetes", "ovnkube-master")))
	g.Expect(objs).To(ContainElement(HaveKubernetesID("DaemonSet", "openshift-ovn-kubernetes", "ovnkube-node")))
	for _, obj := range objs {
		if obj.GetKind() != "Deployment" {
			continue
		}
		sel, found, err := uns.NestedStringMap(obj.Object, "spec", "template", "spec", "nodeSelector")
		g.Expect(err).NotTo(HaveOccurred())
		g.Expect(found).To(BeTrue())
		_, ok := sel["node-role.kubernetes.io/master"]
		g.Expect(ok).To(BeTrue())
	}
	for _, obj := range objs {
		g.Expect(apply.IsObjectSupported(obj)).NotTo(HaveOccurred())
		cur := obj.DeepCopy()
		upd := obj.DeepCopy()
		err = apply.MergeObjectForUpdate(cur, upd)
		g.Expect(err).NotTo(HaveOccurred())
		tweakMetaForCompare(cur)
		g.Expect(cur).To(Equal(upd))
	}
}
func TestFillOVNKubernetesDefaults(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	crd := OVNKubernetesConfig.DeepCopy()
	conf := &crd.Spec
	conf.DefaultNetwork.OVNKubernetesConfig = nil
	m := uint32(8900)
	expected := operv1.NetworkSpec{ServiceNetwork: []string{"172.30.0.0/16"}, ClusterNetwork: []operv1.ClusterNetworkEntry{{CIDR: "10.128.0.0/15", HostPrefix: 23}, {CIDR: "10.0.0.0/14", HostPrefix: 24}}, DefaultNetwork: operv1.DefaultNetworkDefinition{Type: operv1.NetworkTypeOVNKubernetes, OVNKubernetesConfig: &operv1.OVNKubernetesConfig{MTU: &m}}}
	fillOVNKubernetesDefaults(conf, nil, 9000)
	g.Expect(conf).To(Equal(&expected))
}
func TestValidateOVNKubernetes(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	crd := OVNKubernetesConfig.DeepCopy()
	config := &crd.Spec
	ovnConfig := config.DefaultNetwork.OVNKubernetesConfig
	err := validateOVNKubernetes(config)
	g.Expect(err).To(BeEmpty())
	FillDefaults(config, nil)
	errExpect := func(substr string) {
		t.Helper()
		g.Expect(validateOVNKubernetes(config)).To(ContainElement(MatchError(ContainSubstring(substr))))
	}
	mtu := uint32(70000)
	ovnConfig.MTU = &mtu
	errExpect("invalid MTU 70000")
	config.ClusterNetwork = nil
	errExpect("ClusterNetworks cannot be empty")
}
func TestOVNKubernetesIsSafe(t *testing.T) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	g := NewGomegaWithT(t)
	prev := OVNKubernetesConfig.Spec.DeepCopy()
	FillDefaults(prev, nil)
	next := OVNKubernetesConfig.Spec.DeepCopy()
	FillDefaults(next, nil)
	errs := isOVNKubernetesChangeSafe(prev, next)
	g.Expect(errs).To(BeEmpty())
	mtu := uint32(70000)
	next.DefaultNetwork.OVNKubernetesConfig.MTU = &mtu
	errs = isOVNKubernetesChangeSafe(prev, next)
	g.Expect(errs).To(HaveLen(1))
	g.Expect(errs[0]).To(MatchError("cannot change ovn-kubernetes configuration"))
}
