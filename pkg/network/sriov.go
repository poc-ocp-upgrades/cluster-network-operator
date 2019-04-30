package network

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	operv1 "github.com/openshift/api/operator/v1"
	"github.com/openshift/cluster-network-operator/pkg/render"
	"github.com/pkg/errors"
	uns "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type NetConfSRIOV struct {
	Type string `json:"type"`
}

func isOpenShiftSRIOV(conf *operv1.AdditionalNetworkDefinition) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	cni := NetConfSRIOV{}
	err := json.Unmarshal([]byte(conf.RawCNIConfig), &cni)
	if err != nil {
		log.Printf("WARNING: Could not determine if network %q is SR-IOV: %v", conf.Name, err)
		return false
	}
	return cni.Type == "sriov"
}
func renderOpenShiftSRIOV(conf *operv1.AdditionalNetworkDefinition, manifestDir string) ([]*uns.Unstructured, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	var err error
	objs := []*uns.Unstructured{}
	data := render.MakeRenderData()
	data.Data["ReleaseVersion"] = os.Getenv("RELEASE_VERSION")
	data.Data["AdditionalNetworkName"] = conf.Name
	data.Data["AdditionalNetworkNamespace"] = conf.Namespace
	data.Data["AdditionalNetworkConfig"] = conf.RawCNIConfig
	data.Data["SRIOVCNIImage"] = os.Getenv("SRIOV_CNI_IMAGE")
	data.Data["SRIOVDevicePluginImage"] = os.Getenv("SRIOV_DEVICE_PLUGIN_IMAGE")
	objs, err = render.RenderDir(filepath.Join(manifestDir, "network/additional-networks/sriov"), &data)
	if err != nil {
		return nil, errors.Wrap(err, "failed to render OpenShiftSRIOV Network manifests")
	}
	return objs, nil
}
