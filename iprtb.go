package iprtb

import (
	"net"
	"sync"

	"github.com/moznion/go-optional"
)

type routes map[string]RouteEntry

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
		routes: make(routes),
	}
}

func (rt *RouteTable) AddRoute(destination net.IPNet, gateway net.IP, nwInterface string, metric int) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.routes[destination.String()] = RouteEntry{
		Destination: destination,
		Gateway:     gateway,
		NwInterface: nwInterface,
		Metric:      metric,
	}
}

func (rt *RouteTable) MatchRoute(target net.IP) optional.Option[RouteEntry] {
	var matched RouteEntry
	var everMatched bool

	for _, r := range rt.routes { // FIXME no liner-search
		if r.Destination.Contains(target) {
			if !everMatched {
				matched = r
				everMatched = true
				continue
			}

			matchedMaskLen, _ := matched.Destination.Mask.Size()
			newRouteMaskLen, _ := r.Destination.Mask.Size()
			if newRouteMaskLen > matchedMaskLen { // for longest match
				matched = r
			}
		}
	}

	if !everMatched {
		return optional.None[RouteEntry]()
	}
	return optional.Some[RouteEntry](matched)
}

func (rt *RouteTable) RemoveRoute(destination net.IPNet) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	delete(rt.routes, destination.String())
}
