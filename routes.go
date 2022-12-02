package iprtb

import (
	"encoding/json"
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
//
// This struct type supports JSON marshalling and unmarshalling. Please refer also to RouteJSON for more information about that.
type Route struct {
	Destination      *net.IPNet
	Gateway          net.IP
	NetworkInterface string
	Metric           int
}

func (r Route) String() string {
	return fmt.Sprintf("%s\t%s\t%s\t%d", r.Destination.String(), r.Gateway.String(), r.NetworkInterface, r.Metric)
}

// RouteJSON is an intermediate representation for Route to do JSON marshalling and unmarshalling.
//
// When it marshals Route to JSON bytes, it transforms a Route value to a RouteJSON and applies that to `json.Marshal()`.
// Also, hen it attempts to unmarshal JSON bytes to a Route value, it applies `json.Unmarshal()` to that JSON bytes to
// derive the RouteJSON value at first, and after that, it converts that RouteJSON value into a Route value.
type RouteJSON struct {
	Destination      string `json:"destination"`
	Gateway          string `json:"gateway"`
	NetworkInterface string `json:"networkInterface"`
	Metric           int    `json:"metric"`
}

func (r Route) MarshalJSON() ([]byte, error) {
	rj := &RouteJSON{
		Destination:      r.Destination.String(),
		Gateway:          r.Gateway.String(),
		NetworkInterface: r.NetworkInterface,
		Metric:           r.Metric,
	}
	return json.Marshal(rj)
}

func (r *Route) UnmarshalJSON(data []byte) error {
	var rj RouteJSON
	err := json.Unmarshal(data, &rj)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Route: %w", err)
	}

	_, destination, err := net.ParseCIDR(rj.Destination)
	if err != nil {
		return fmt.Errorf(`failed to unmarshal Route; it cannot parse the value of "destination" property as net.IPNet: %w`, err)
	}

	r.Destination = destination
	r.Gateway = net.ParseIP(rj.Gateway)
	r.NetworkInterface = rj.NetworkInterface
	r.Metric = rj.Metric
	return nil
}
