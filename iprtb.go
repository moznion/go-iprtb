package iprtb

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/moznion/go-optional"
)

// ErrInvalidIPv6Length represents the error that indicates given IPv6 address has invalid length.
var ErrInvalidIPv6Length = errors.New("given IPv6 address doesn't satisfy the IPv6 length")

// RouteTable is a route table implementation.
type RouteTable struct {
	routes            *node
	label2Destination map[string]*net.IPNet
	mu                sync.Mutex
}

// RouteEntry is an entry of route table.
type RouteEntry struct {
	Destination net.IPNet
	Gateway     net.IP
	NwInterface string
	Metric      int
}

// NewRouteTable makes a new RouteTable value.
func NewRouteTable() *RouteTable {
	return &RouteTable{
		routes:            &node{},
		label2Destination: map[string]*net.IPNet{},
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
	return rt.addRoute(destination, gateway, nwInterface, metric)
}

func (rt *RouteTable) AddRouteWithLabel(destination net.IPNet, gateway net.IP, nwInterface string, metric int, label string) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	err := rt.addRoute(destination, gateway, nwInterface, metric)
	if err != nil {
		return err
	}
	rt.label2Destination[label] = &destination
	return nil
}

func (rt *RouteTable) UpdateByLabel(label string, gateway net.IP, nwInterface string, metric int) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	destination := rt.label2Destination[label]
	if destination == nil {
		return nil
	}
	return rt.addRoute(*destination, gateway, nwInterface, metric)
}

func (rt *RouteTable) addRoute(destination net.IPNet, gateway net.IP, nwInterface string, metric int) error {
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
	return rt.removeRoute(destination)
}

func (rt *RouteTable) RemoveRouteByLabel(label string) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	destination := rt.label2Destination[label]
	if destination == nil {
		return nil
	}

	err := rt.removeRoute(*destination)
	if err != nil {
		return err
	}
	delete(rt.label2Destination, label)
	return nil
}

func (rt *RouteTable) removeRoute(destination net.IPNet) error {
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

func (rt *RouteTable) DumpRouteTable() []*RouteEntry {
	return rt.scanNode(rt.routes)
}

func (rt *RouteTable) scanNode(visitNode *node) []*RouteEntry {
	routeEntries := make([]*RouteEntry, 0)

	if visitNode.zeroBitNode != nil {
		routeEntries = append(routeEntries, rt.scanNode(visitNode.zeroBitNode)...)
	}
	if visitNode.oneBitNode != nil {
		routeEntries = append(routeEntries, rt.scanNode(visitNode.oneBitNode)...)
	}

	if visitNode.routeEntry != nil {
		routeEntries = append(routeEntries, visitNode.routeEntry)
	}

	return routeEntries
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
