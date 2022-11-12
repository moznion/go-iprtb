package iprtb

import (
	"net"
	"sync"

	"github.com/moznion/go-optional"
)

type routes []RouteEntry

type RouteTable struct {
	routes routes
	mu     sync.Mutex
}

type RouteEntry struct {
	Destination net.IPNet
	Gateway     net.IP
	NwInterface string
	Metric      int
}

func NewRouteTable() *RouteTable {
	return &RouteTable{
		routes: make(routes, 0),
	}
}

func (rt *RouteTable) AddRoute(destination net.IPNet, gateway net.IP, nwInterface string, metric int) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.routes = append(rt.routes, RouteEntry{
		Destination: destination,
		Gateway:     gateway,
		NwInterface: nwInterface,
		Metric:      metric,
	})
}

func (rt *RouteTable) MatchRoute(destination net.IP) optional.Option[RouteEntry] {
	var matched RouteEntry
	var everMatched bool

	for _, r := range rt.routes { // FIXME no liner-search
		if r.Destination.Contains(destination) {
			if !everMatched {
				matched = r
				everMatched = true
				continue
			}

			matchedMaskLen, _ := matched.Destination.Mask.Size()
			newRouteMaskLen, _ := r.Destination.Mask.Size()
			if newRouteMaskLen > matchedMaskLen {
				matched = r
			}
		}
	}

	if !everMatched {
		return optional.None[RouteEntry]()
	}
	return optional.Some[RouteEntry](matched)
}
