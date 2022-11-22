package iprtb

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"

	"github.com/moznion/go-optional"
)

// ErrInvalidIPv6Length represents the error that indicates given IPv6 address has invalid length.
var ErrInvalidIPv6Length = errors.New("given IPv6 address doesn't satisfy the IPv6 length")

// RouteTable is a routing table implementation.
type RouteTable struct {
	routes            *node
	label2Destination map[string]*net.IPNet
	destination2Label map[string]string
	mu                sync.Mutex
}

// NewRouteTable makes a new RouteTable value.
func NewRouteTable() *RouteTable {
	return &RouteTable{
		routes:            &node{},
		label2Destination: map[string]*net.IPNet{},
		destination2Label: map[string]string{},
	}
}

// AddRoute adds a route to the routing table.
// If the destination has already existed in the routing table, this overwrites the route information by the given route.
// In other words, this function behaves as well as "update" against the existing routes.
func (rt *RouteTable) AddRoute(ctx context.Context, route *Route) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	return rt.addRoute(ctx, route)
}

// AddRouteWithLabel adds a route to the routing table with a label.
// If the destination has already existed in the routing table, this overwrites the route information by the given route.
// The label is capable to use by UpdateRouteByLabel and RemoveRouteByLabel functions instead of the actual destination information.
// If there already had the given label, it overwrites by the given one.
func (rt *RouteTable) AddRouteWithLabel(ctx context.Context, label string, route *Route) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	err := rt.addRoute(ctx, route)
	if err != nil {
		return err
	}
	rt.label2Destination[label] = route.Destination
	rt.destination2Label[route.Destination.String()] = label
	return nil
}

// UpdateRouteByLabel updates the existing route that is associated with the label by given parameters.
// If there is no route that is associated with a given label, this function does nothing.
func (rt *RouteTable) UpdateRouteByLabel(ctx context.Context, label string, gateway net.IP, nwInterface string, metric int) error {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	destination := rt.label2Destination[label]
	if destination == nil {
		return nil
	}
	return rt.addRoute(ctx, &Route{
		Destination:      destination,
		Gateway:          gateway,
		NetworkInterface: nwInterface,
		Metric:           metric,
	})
}

func (rt *RouteTable) addRoute(ctx context.Context, route *Route) error {
	destination := route.Destination
	terminalRoute := &Route{
		Destination:      destination,
		Gateway:          route.Gateway,
		NetworkInterface: route.NetworkInterface,
		Metric:           route.Metric,
	}

	currentNode := rt.routes // root
	maskLen, _ := destination.Mask.Size()
	if maskLen <= 0 {
		currentNode.route = terminalRoute
		return nil
	}

	dstIP, err := adjustIPLength(destination.IP)
	if err != nil {
		return fmt.Errorf("failed to add a route: %w", err)
	}

	for _, b := range dstIP {
		for rightShift := 0; rightShift <= 7; rightShift++ {
			bit := toBit(b, rightShift)

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

			maskLen--
			if maskLen <= 0 { // terminated
				nextNode.route = terminalRoute
				return nil
			}

			currentNode = nextNode
		}
	}

	return nil
}

// RemoveRoute removes a route that is associated with a given destination. This returns the removed route information that is wrapped by optional.
// If there is no route to remove, this does nothing and returns `None` as the removed route.
func (rt *RouteTable) RemoveRoute(ctx context.Context, destination *net.IPNet) (optional.Option[Route], error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	maybeRemovedRoute, err := rt.removeRoute(ctx, destination)
	if err != nil {
		return optional.None[Route](), err
	}
	if label, ok := rt.destination2Label[destination.String()]; ok {
		delete(rt.destination2Label, destination.String())
		delete(rt.label2Destination, label)
	}
	return maybeRemovedRoute, nil
}

// RemoveRouteByLabel removes a route that is associated with a given label, instead of the actual destination information. This returns the removed route information that is wrapped by optional.
// If there is no route that is associated with a given label or the actual destination, this function does nothing and returns `None` as the removed route.
func (rt *RouteTable) RemoveRouteByLabel(ctx context.Context, label string) (optional.Option[Route], error) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	destination := rt.label2Destination[label]
	if destination == nil {
		return nil, nil
	}

	maybeRemovedRoute, err := rt.removeRoute(ctx, destination)
	if err != nil {
		return optional.None[Route](), err
	}
	delete(rt.label2Destination, label)
	delete(rt.destination2Label, destination.String())
	return maybeRemovedRoute, nil
}

func (rt *RouteTable) removeRoute(ctx context.Context, destination *net.IPNet) (optional.Option[Route], error) {
	prevNode := rt.routes // root
	maskLen, _ := destination.Mask.Size()
	if maskLen <= 0 {
		removedRoute := optional.FromNillable[Route](prevNode.route)
		prevNode.route = nil
		return removedRoute, nil
	}

	dstIP, err := adjustIPLength(destination.IP)
	if err != nil {
		return optional.None[Route](), fmt.Errorf("failed to remove a route: %w", err)
	}

	var pathNodes []**node // to brune the not terminated branch

	for _, b := range dstIP {
		for rightShift := 0; rightShift <= 7; rightShift++ {
			bit := toBit(b, rightShift)

			maskLen--

			var nextNode **node
			if bit == 0 {
				nextNode = &prevNode.zeroBitNode
			} else {
				nextNode = &prevNode.oneBitNode
			}

			if *nextNode == nil {
				// there is no terminal route that is associated with the given destination to remove; do nothing
				return optional.None[Route](), nil
			}

			if maskLen <= 0 {
				removedRoute := optional.None[Route]()
				if (*nextNode).route != nil {
					removedRoute = optional.Some[Route](*((*nextNode).route))
					// node terminated: should remove a route
					if (*nextNode).zeroBitNode == nil && (*nextNode).oneBitNode == nil {
						// this terminal node doesn't have any children; do pruning including the terminal node itself
						// NOTE: it must apply the node pruning processing as reverse order (i.e. the direction from child to parent).
						*nextNode = nil

						pathNodesLen := len(pathNodes)
						for i := pathNodesLen - 1; i >= 0; i-- {
							pathNode := pathNodes[i]
							if (*pathNode).zeroBitNode == nil && (*pathNode).oneBitNode == nil {
								*pathNode = nil
							} else {
								// if the currently processed node isn't removed, the all following nodes (i.e. ancestor nodes) have at least one child, so it's okay to break the processing here.
								break
							}
						}
					} else {
						// this node has some children, so it removes only routing info
						(*nextNode).route = nil
					}
				}
				return removedRoute, nil
			}

			if (*nextNode).route == nil {
				// path (i.e. not a terminal) node: this is a candidate to pruning
				pathNodes = append(pathNodes, nextNode)
			} else {
				// terminal node: it mustn't prune the branch that has this node and the ancestor nodes
				pathNodes = []**node{}
			}

			prevNode = *nextNode
		}
	}

	return optional.None[Route](), nil
}

// ClearRoutes removes all routes from the routing table.
func (rt *RouteTable) ClearRoutes(ctx context.Context) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.routes = &node{}
	rt.label2Destination = map[string]*net.IPNet{}
	rt.destination2Label = map[string]string{}
}

// MatchRoute attempts to check whether the given IP address matches the routing table or not.
// If there is matched route, this returns that route information that is wrapped by optional.Some.
// Else, this returns the value of optional.None.
func (rt *RouteTable) MatchRoute(ctx context.Context, target net.IP) (optional.Option[Route], error) {
	target, err := adjustIPLength(target)
	if err != nil {
		return optional.None[Route](), fmt.Errorf("invalid target IP address on matching a route => %s: %w", target, err)
	}

	var matchedRoute *Route
	visitNode := rt.routes
	for _, b := range target {
		for rightShift := 0; rightShift <= 7; rightShift++ {
			if visitNode.route != nil {
				matchedRoute = visitNode.route
			}

			bit := toBit(b, rightShift)
			if bit == 0 {
				if visitNode.zeroBitNode == nil {
					return optional.FromNillable[Route](matchedRoute), nil
				}
				visitNode = visitNode.zeroBitNode
			} else {
				if visitNode.oneBitNode == nil {
					return optional.FromNillable[Route](matchedRoute), nil
				}
				visitNode = visitNode.oneBitNode
			}
		}
	}
	if visitNode.route != nil {
		matchedRoute = visitNode.route
	}

	return optional.FromNillable[Route](matchedRoute), nil
}

// FindRoute attempts to find the route information that is matched with the given IP address.
// If that route is found this returns true.
// This function doesn't respect the longest match, so the performance of this function would be better than MatchRoute but this doesn't return the actual detailed information.
func (rt *RouteTable) FindRoute(ctx context.Context, target net.IP) (bool, error) {
	target, err := adjustIPLength(target)
	if err != nil {
		return false, fmt.Errorf("invalid target IP address on finding a route => %s: %w", target, err)
	}

	visitNode := rt.routes
	for _, b := range target {
		for rightShift := 0; rightShift <= 7; rightShift++ {
			if visitNode.route != nil {
				return true, nil
			}

			bit := toBit(b, rightShift)
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
	return visitNode.route != nil, nil
}

// DumpRouteTable dumps the configurations of the routing table.
// The result value supports String() method so that be able to do stringify.
func (rt *RouteTable) DumpRouteTable(ctx context.Context) Routes {
	return rt.scanNode(rt.routes)
}

func (rt *RouteTable) scanNode(visitNode *node) Routes {
	routes := make([]*Route, 0)

	if visitNode.zeroBitNode != nil {
		routes = append(routes, rt.scanNode(visitNode.zeroBitNode)...)
	}
	if visitNode.oneBitNode != nil {
		routes = append(routes, rt.scanNode(visitNode.oneBitNode)...)
	}

	if visitNode.route != nil {
		routes = append(routes, visitNode.route)
	}

	return routes
}

func toBit(b byte, rightShift int) byte {
	mask := byte(0b10000000 >> rightShift)
	return byte((b & mask) >> (7 - rightShift))
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

type node struct {
	zeroBitNode *node
	oneBitNode  *node
	route       *Route
}
