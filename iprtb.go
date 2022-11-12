package iprtb

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/moznion/go-optional"
)

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

func (rt *RouteTable) AddRoute(destination net.IPNet, gateway net.IP, nwInterface string, metric int) {
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
		return
	}

	dstIP, _ := adjustIPLength(destination.IP)
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

				return
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
}

func (rt *RouteTable) MatchRoute(target net.IP) (optional.Option[RouteEntry], error) {
	var matchedRoute *RouteEntry
	visitNode := rt.routes

	target, err := adjustIPLength(target)
	if err != nil {
		return optional.None[RouteEntry](), fmt.Errorf("invalid target IP address => %s: %w", target, err)
	}

iploop:
	for _, b := range target {
		for rshift := 0; rshift <= 7; rshift++ {
			if visitNode.routeEntry != nil {
				matchedRoute = visitNode.routeEntry
			}

			bit := toBit(b, rshift)
			if bit == 0 {
				if visitNode.zeroBitNode == nil {
					break iploop
				}
				visitNode = visitNode.zeroBitNode
			} else {
				if visitNode.oneBitNode == nil {
					break iploop
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

func (rt *RouteTable) RemoveRoute(destination net.IPNet) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	currentNode := rt.routes // root
	maskLen, _ := destination.Mask.Size()
	if maskLen <= 0 {
		currentNode.routeEntry = nil
		return
	}

	dstIP, _ := adjustIPLength(destination.IP)
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
				return
			}

			var nextNode *node
			if bit == 0 {
				nextNode = currentNode.zeroBitNode
			} else {
				nextNode = currentNode.oneBitNode
			}
			if nextNode == nil {
				return
			}

			currentNode = nextNode
		}
	}
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
		return nil, errors.New("given IPv6 address doesn't satisfy the IPv6 length")
	}
	return givenIP, nil
}
