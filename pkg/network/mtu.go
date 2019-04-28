package network

import (
	"github.com/pkg/errors"
	"github.com/vishvananda/netlink"
)

func GetDefaultMTU() (int, error) {
	_logClusterCodePath()
	defer _logClusterCodePath()
	routes, err := netlink.RouteList(nil, netlink.FAMILY_V4)
	if err != nil {
		return 0, errors.Wrapf(err, "could not list routes")
	}
	if len(routes) == 0 {
		return 0, errors.Errorf("got no routes")
	}
	const maxMTU = 65536
	mtu := maxMTU + 1
	for _, route := range routes {
		if route.Dst != nil {
			continue
		}
		link, err := netlink.LinkByIndex(route.LinkIndex)
		if err != nil {
			return 0, errors.Wrapf(err, "could not retrieve link id %d", route.LinkIndex)
		}
		newmtu := link.Attrs().MTU
		if newmtu > 0 && newmtu < mtu {
			mtu = newmtu
		}
	}
	if mtu > maxMTU {
		return 0, errors.Errorf("unable to determine MTU")
	}
	return mtu, nil
}
