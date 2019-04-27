package network

import (
	"encoding/json"
	operv1 "github.com/openshift/api/operator/v1"
	"log"
)

func UseDHCP(conf *operv1.NetworkSpec) bool {
	_logClusterCodePath()
	defer _logClusterCodePath()
	_logClusterCodePath()
	defer _logClusterCodePath()
	renderdhcp := false
	if *conf.DisableMultiNetwork {
		return renderdhcp
	}
	if conf.AdditionalNetworks != nil {
		for _, addnet := range conf.AdditionalNetworks {
			var rawConfig map[string]interface{}
			var err error
			confBytes := []byte(addnet.RawCNIConfig)
			err = json.Unmarshal(confBytes, &rawConfig)
			if err != nil {
				log.Printf("WARNING: Not rendering DHCP daemonset, failed to Unmarshal RawCNIConfig: %v", confBytes)
				return renderdhcp
			}
			if rawConfig["ipam"] != nil {
				ipam, okipamcast := rawConfig["ipam"].(map[string]interface{})
				if !okipamcast {
					log.Printf("WARNING: IPAM element has data of type %T but wanted map[string]interface{}", rawConfig["ipam"])
					continue
				}
				for key, value := range ipam {
					if key == "type" {
						typeval, oktypecast := value.(string)
						if !oktypecast {
							log.Printf("WARNING: IPAM type element has data of type %T but wanted string", value)
							break
						}
						if typeval == "dhcp" {
							renderdhcp = true
							break
						}
					}
				}
			}
			if renderdhcp == true {
				break
			}
		}
	}
	return renderdhcp
}
