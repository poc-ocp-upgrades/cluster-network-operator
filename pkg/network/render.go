package network

import (
	"log"
	"net"
	"reflect"
	"strings"
	"github.com/pkg/errors"
	operv1 "github.com/openshift/api/operator/v1"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Render(conf *operv1.NetworkSpec, manifestDir string) ([]*uns.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	log.Printf("Starting render phase")
	objs := []*uns.Unstructured{}
	o, err := RenderMultus(conf, manifestDir)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)
	o, err = RenderDefaultNetwork(conf, manifestDir)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)
	o, err = RenderAdditionalNetworks(conf, manifestDir)
	if err != nil {
		return nil, err
	}
	objs = append(objs, o...)
	log.Printf("Render phase done, rendered %d objects", len(objs))
	return objs, nil
}
func Canonicalize(conf *operv1.NetworkSpec) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch strings.ToLower(string(conf.DefaultNetwork.Type)) {
	case strings.ToLower(string(operv1.NetworkTypeOpenShiftSDN)):
		conf.DefaultNetwork.Type = operv1.NetworkTypeOpenShiftSDN
	case strings.ToLower(string(operv1.NetworkTypeOVNKubernetes)):
		conf.DefaultNetwork.Type = operv1.NetworkTypeOVNKubernetes
	}
	if conf.DefaultNetwork.Type == operv1.NetworkTypeOpenShiftSDN && conf.DefaultNetwork.OpenShiftSDNConfig != nil {
		sdnc := conf.DefaultNetwork.OpenShiftSDNConfig
		switch strings.ToLower(string(sdnc.Mode)) {
		case strings.ToLower(string(operv1.SDNModeMultitenant)):
			sdnc.Mode = operv1.SDNModeMultitenant
		case strings.ToLower(string(operv1.SDNModeNetworkPolicy)):
			sdnc.Mode = operv1.SDNModeNetworkPolicy
		case strings.ToLower(string(operv1.SDNModeSubnet)):
			sdnc.Mode = operv1.SDNModeSubnet
		}
	}
	for _, an := range conf.AdditionalNetworks {
		switch strings.ToLower(string(an.Type)) {
		case strings.ToLower(string(operv1.NetworkTypeRaw)):
			an.Type = operv1.NetworkTypeRaw
		}
	}
}
func Validate(conf *operv1.NetworkSpec) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	errs := []error{}
	errs = append(errs, ValidateIPPools(conf)...)
	errs = append(errs, ValidateDefaultNetwork(conf)...)
	errs = append(errs, ValidateMultus(conf)...)
	if len(errs) > 0 {
		return errors.Errorf("invalid configuration: %v", errs)
	}
	return nil
}
func FillDefaults(conf, previous *operv1.NetworkSpec) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	hostMTU, err := GetDefaultMTU()
	if hostMTU == 0 {
		hostMTU = 1500
	}
	if previous == nil {
		if err != nil {
			log.Printf("Failed MTU probe, failling back to 1500: %v", err)
		} else {
			log.Printf("Detected uplink MTU %d", hostMTU)
		}
	}
	if conf.DisableMultiNetwork == nil {
		disable := false
		conf.DisableMultiNetwork = &disable
	}
	FillDefaultNetworkDefaults(conf, previous, hostMTU)
}
func IsChangeSafe(prev, next *operv1.NetworkSpec) error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if prev == nil {
		return nil
	}
	if reflect.DeepEqual(prev, next) {
		return nil
	}
	errs := []error{}
	if !reflect.DeepEqual(prev.ClusterNetwork, next.ClusterNetwork) {
		errs = append(errs, errors.Errorf("cannot change ClusterNetworks"))
	}
	if !reflect.DeepEqual(prev.ServiceNetwork, next.ServiceNetwork) {
		errs = append(errs, errors.Errorf("cannot change ServiceNetwork"))
	}
	errs = append(errs, IsDefaultNetworkChangeSafe(prev, next)...)
	if *prev.DisableMultiNetwork != *next.DisableMultiNetwork {
		errs = append(errs, errors.Errorf("cannot change DisableMultiNetwork"))
	}
	if len(errs) > 0 {
		return errors.Errorf("invalid configuration: %v", errs)
	}
	return nil
}
func ValidateIPPools(conf *operv1.NetworkSpec) []error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	errs := []error{}
	for idx, pool := range conf.ClusterNetwork {
		_, _, err := net.ParseCIDR(pool.CIDR)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "could not parse ClusterNetwork %d CIDR %q", idx, pool.CIDR))
		}
	}
	for idx, pool := range conf.ServiceNetwork {
		_, _, err := net.ParseCIDR(pool)
		if err != nil {
			errs = append(errs, errors.Wrapf(err, "could not parse ServiceNetwork %d CIDR %q", idx, pool))
		}
	}
	return errs
}
func ValidateMultus(conf *operv1.NetworkSpec) []error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	deployMultus := true
	if conf.DisableMultiNetwork != nil && *conf.DisableMultiNetwork {
		deployMultus = false
	}
	if !deployMultus && len(conf.AdditionalNetworks) > 0 {
		return []error{errors.Errorf("additional networks cannot be specified without deploying Multus")}
	}
	return []error{}
}
func ValidateDefaultNetwork(conf *operv1.NetworkSpec) []error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch conf.DefaultNetwork.Type {
	case operv1.NetworkTypeOpenShiftSDN:
		return validateOpenShiftSDN(conf)
	case operv1.NetworkTypeOVNKubernetes:
		return validateOVNKubernetes(conf)
	default:
		return nil
	}
}
func RenderDefaultNetwork(conf *operv1.NetworkSpec, manifestDir string) ([]*uns.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	dn := conf.DefaultNetwork
	if errs := ValidateDefaultNetwork(conf); len(errs) > 0 {
		return nil, errors.Errorf("invalid Default Network configuration: %v", errs)
	}
	switch dn.Type {
	case operv1.NetworkTypeOpenShiftSDN:
		return renderOpenShiftSDN(conf, manifestDir)
	case operv1.NetworkTypeOVNKubernetes:
		return renderOVNKubernetes(conf, manifestDir)
	default:
		log.Printf("NOTICE: Unknown network type %s, ignoring", dn.Type)
		return nil, nil
	}
}
func FillDefaultNetworkDefaults(conf, previous *operv1.NetworkSpec, hostMTU int) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	switch conf.DefaultNetwork.Type {
	case operv1.NetworkTypeOpenShiftSDN:
		fillOpenShiftSDNDefaults(conf, previous, hostMTU)
	case operv1.NetworkTypeOVNKubernetes:
		fillOVNKubernetesDefaults(conf, previous, hostMTU)
	default:
	}
}
func IsDefaultNetworkChangeSafe(prev, next *operv1.NetworkSpec) []error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if prev.DefaultNetwork.Type != next.DefaultNetwork.Type {
		return []error{errors.Errorf("cannot change default network type")}
	}
	switch prev.DefaultNetwork.Type {
	case operv1.NetworkTypeOpenShiftSDN:
		return isOpenShiftSDNChangeSafe(prev, next)
	case operv1.NetworkTypeOVNKubernetes:
		return isOVNKubernetesChangeSafe(prev, next)
	default:
		return nil
	}
}
func ValidateAdditionalNetworks(conf *operv1.NetworkSpec) [][]error {
	_logClusterCodePath()
	defer _logClusterCodePath()
	out := [][]error{}
	ans := conf.AdditionalNetworks
	for _, an := range ans {
		switch an.Type {
		case operv1.NetworkTypeRaw:
			if errs := validateRaw(&an); len(errs) > 0 {
				out = append(out, errs)
			}
		default:
			out = append(out, []error{errors.Errorf("unknown or unsupported NetworkType: %s", an.Type)})
		}
	}
	return out
}
func RenderAdditionalNetworks(conf *operv1.NetworkSpec, manifestDir string) ([]*uns.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var err error
	ans := conf.AdditionalNetworks
	out := []*uns.Unstructured{}
	objs := []*uns.Unstructured{}
	if errs := ValidateAdditionalNetworks(conf); len(errs) > 0 {
		return nil, errors.Errorf("invalid Additional Network Configuration: %v", errs)
	}
	if len(ans) == 0 {
		return nil, nil
	}
	for _, an := range ans {
		switch an.Type {
		case operv1.NetworkTypeRaw:
			if isOpenShiftSRIOV(&an) {
				objs, err = renderOpenShiftSRIOV(&an, manifestDir)
			} else {
				objs, err = renderRawCNIConfig(&an, manifestDir)
			}
			if err != nil {
				return nil, err
			}
			out = append(out, objs...)
		default:
			return nil, errors.Errorf("unknown or unsupported NetworkType: %s", an.Type)
		}
	}
	return out, nil
}
func RenderMultus(conf *operv1.NetworkSpec, manifestDir string) ([]*uns.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	if *conf.DisableMultiNetwork {
		return nil, nil
	}
	var err error
	out := []*uns.Unstructured{}
	objs := []*uns.Unstructured{}
	objs, err = renderAdditionalNetworksCRD(manifestDir)
	if err != nil {
		return nil, err
	}
	out = append(out, objs...)
	usedhcp := UseDHCP(conf)
	objs, err = renderMultusConfig(manifestDir, usedhcp)
	if err != nil {
		return nil, err
	}
	out = append(out, objs...)
	return out, nil
}
