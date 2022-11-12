package iprtb

import (
	"fmt"
	"net"
)

// Routes is a list of Route
type Routes []*Route

func (rs Routes) String() string {
	str := ""
	for _, r := range rs {
		str += r.String() + "\n"
	}
	return str
}

// Route is an entry of routing table.
type Route struct {
	Destination      net.IPNet
	Gateway          net.IP
	NetworkInterface string
	Metric           int
}

func (r Route) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d", r.Destination.String(), r.Gateway.String(), r.NetworkInterface, r.Metric)
}
