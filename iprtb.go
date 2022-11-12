package iprtb

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/moznion/go-optional"
)

var ErrInvalidIPv6Length = errors.New("given IPv6 address doesn't satisfy the IPv6 length")

type RouteTable struct {
	routes *node
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
		routes: &node{},
	}
}

type node struct {
	zeroBitNode *node
	oneBitNode  *node
	routeEntry  *RouteEntry
}

func (rt *RouteTable) AddRoute(destination net.IPNet, gateway net.IP, nwInterface string, metric int) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	terminalRouteEntry := &RouteEntry{
		Destination: destination,
		Gateway:     gateway,
		NwInterface: nwInterface,
		Metric:      metric,
	}

	currentNode := rt.routes // root
	maskLen, _ := destination.Mask.Size()
	if maskLen <= 0 {
		currentNode.routeEntry = terminalRouteEntry
		return nil
	}

	dstIP, err := adjustIPLength(destination.IP)
	if err != nil {
		return fmt.Errorf("failed to add a route: %w", err)
	}

	for _, b := range dstIP {
		for rshift := 0; rshift <= 7; rshift++ {
			bit := toBit(b, rshift)

			maskLen--
			if maskLen <= 0 { // terminated
				if bit == 0 {
					if currentNode.zeroBitNode == nil {
						currentNode.zeroBitNode = &node{}
					}
					currentNode.zeroBitNode.routeEntry = terminalRouteEntry
				} else {
					if currentNode.oneBitNode == nil {
						currentNode.oneBitNode = &node{}
					}
					currentNode.oneBitNode.routeEntry = terminalRouteEntry
				}

				return nil
			}

			var nextNode *node
			if bit == 0 {
				if currentNode.zeroBitNode == nil {
					currentNode.zeroBitNode = &node{}
				}
				nextNode = currentNode.zeroBitNode
			} else {
				if currentNode.oneBitNode == nil {
					currentNode.oneBitNode = &node{}
				}
				nextNode = currentNode.oneBitNode
			}
			currentNode = nextNode
		}
	}

	return nil
}

func (rt *RouteTable) RemoveRoute(destination net.IPNet) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	currentNode := rt.routes // root
	maskLen, _ := destination.Mask.Size()
	if maskLen <= 0 {
		currentNode.routeEntry = nil
		return nil
	}

	dstIP, err := adjustIPLength(destination.IP)
	if err != nil {
		return fmt.Errorf("failed to remove a route: %w", err)
	}

	for _, b := range dstIP {
		for rshift := 0; rshift <= 7; rshift++ {
			bit := toBit(b, rshift)

			maskLen--
			if maskLen <= 0 { // terminated
				if bit == 0 && currentNode.zeroBitNode != nil {
					currentNode.zeroBitNode.routeEntry = nil
				} else if currentNode.oneBitNode != nil {
					currentNode.oneBitNode.routeEntry = nil
				}
				return nil
			}

			var nextNode *node
			if bit == 0 {
				nextNode = currentNode.zeroBitNode
			} else {
				nextNode = currentNode.oneBitNode
			}
			if nextNode == nil {
				return nil
			}

			currentNode = nextNode
		}
	}

	return nil
}

func (rt *RouteTable) MatchRoute(target net.IP) (optional.Option[RouteEntry], error) {
	target, err := adjustIPLength(target)
	if err != nil {
		return optional.None[RouteEntry](), fmt.Errorf("invalid target IP address on matching a route => %s: %w", target, err)
	}

	var matchedRoute *RouteEntry
	visitNode := rt.routes
	for _, b := range target {
		for rshift := 0; rshift <= 7; rshift++ {
			if visitNode.routeEntry != nil {
				matchedRoute = visitNode.routeEntry
			}

			bit := toBit(b, rshift)
			if bit == 0 {
				if visitNode.zeroBitNode == nil {
					return optional.FromNillable[RouteEntry](matchedRoute), nil
				}
				visitNode = visitNode.zeroBitNode
			} else {
				if visitNode.oneBitNode == nil {
					return optional.FromNillable[RouteEntry](matchedRoute), nil
				}
				visitNode = visitNode.oneBitNode
			}
		}
	}
	if visitNode.routeEntry != nil {
		matchedRoute = visitNode.routeEntry
	}

	return optional.FromNillable[RouteEntry](matchedRoute), nil
}

func (rt *RouteTable) FindRoute(target net.IP) (bool, error) {
	target, err := adjustIPLength(target)
	if err != nil {
		return false, fmt.Errorf("invalid target IP address on finding a route => %s: %w", target, err)
	}

	visitNode := rt.routes
	for _, b := range target {
		for rshift := 0; rshift <= 7; rshift++ {
			if visitNode.routeEntry != nil {
				return true, nil
			}

			bit := toBit(b, rshift)
			if bit == 0 {
				if visitNode.zeroBitNode == nil {
					return false, nil
				}
				visitNode = visitNode.zeroBitNode
			} else {
				if visitNode.oneBitNode == nil {
					return false, nil
				}
				visitNode = visitNode.oneBitNode
			}
		}
	}
	return visitNode.routeEntry != nil, nil
}

func toBit(b byte, rshift int) byte {
	mask := byte(0b10000000 >> rshift)
	return byte((b & mask) >> (7 - rshift))
}

func adjustIPLength(givenIP net.IP) (net.IP, error) {
	if ipv4 := givenIP.To4(); ipv4 != nil {
		return ipv4, nil
	}

	// ipv6
	if len(givenIP) != net.IPv6len {
		return nil, ErrInvalidIPv6Length
	}
	return givenIP, nil
}
