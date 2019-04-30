package network

func GetDefaultMTU() (int, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	return 1500, nil
}
