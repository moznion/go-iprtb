package iprtb

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
)

func ExampleRouteTable_MatchRoute() {
	ctx := context.Background()

	rtb := NewRouteTable()

	err := rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	err = rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 255),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	maybeRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())
	fmt.Println(maybeRoute.Unwrap().String())

	maybeRoute, err = rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 254))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())
	fmt.Println(maybeRoute.Unwrap().String())

	// longest match
	maybeRoute, err = rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())
	fmt.Println(maybeRoute.Unwrap().String())

	// not routes
	maybeRoute, err = rtb.MatchRoute(ctx, net.IPv4(198, 51, 100, 123))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())

	// Output:
	// true
	// 192.0.2.0/24	192.0.2.1	ifb0	1
	// true
	// 192.0.2.0/24	192.0.2.1	ifb0	1
	// true
	// 192.0.2.255/32	192.0.2.255	ifb0	1
	// false
}

func ExampleRouteTable_FindRoute() {
	ctx := context.Background()

	rtb := NewRouteTable()

	err := rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	err = rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 255),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	found, err = rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 254))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// longest match (but this is early broke)
	found, err = rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// not routes
	found, err = rtb.FindRoute(ctx, net.IPv4(198, 51, 100, 123))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// Output:
	// true
	// true
	// true
	// false
}

func ExampleRouteTable_RemoveRoute() {
	ctx := context.Background()

	rtb := NewRouteTable()

	dst := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	err := rtb.AddRoute(ctx, &Route{
		Destination:      dst,
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	maybeRemovedRoute, err := rtb.RemoveRoute(ctx, dst)
	if err != nil {
		panic(err)
	}
	removedRoute := maybeRemovedRoute.Unwrap()
	fmt.Printf("removed route: destination %s, gateway %s, networkInterface %s, metric %d\n", removedRoute.Destination, removedRoute.Gateway, removedRoute.NetworkInterface, removedRoute.Metric)

	// route has been removed
	found, err = rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// Output:
	// true
	// removed route: destination 192.0.2.0/24, gateway 192.0.2.1, networkInterface ifb0, metric 1
	// false
}

func ExampleRouteTable_RemoveRouteByLabel() {
	ctx := context.Background()

	rtb := NewRouteTable()

	label := "__label__"
	err := rtb.AddRouteWithLabel(ctx, label, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	maybeRemovedRoute, err := rtb.RemoveRouteByLabel(ctx, label)
	if err != nil {
		panic(err)
	}
	removedRoute := maybeRemovedRoute.Unwrap()
	fmt.Printf("removed route: destination %s, gateway %s, networkInterface %s, metric %d\n", removedRoute.Destination, removedRoute.Gateway, removedRoute.NetworkInterface, removedRoute.Metric)

	// route has been removed
	found, err = rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// Output:
	// true
	// removed route: destination 192.0.2.0/24, gateway 192.0.2.1, networkInterface ifb0, metric 1
	// false
}

func ExampleRouteTable_UpdateRouteByLabel() {
	ctx := context.Background()

	rtb := NewRouteTable()

	label := "__label__"
	err := rtb.AddRouteWithLabel(ctx, label, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	matched, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.Unwrap().String())

	err = rtb.UpdateRouteByLabel(ctx, label, net.IPv4(192, 0, 2, 2), "ifb1", 2)
	if err != nil {
		panic(err)
	}
	matched, err = rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.Unwrap().String())

	// Output:
	// 192.0.2.0/24	192.0.2.1	ifb0	1
	// 192.0.2.0/24	192.0.2.2	ifb1	2
}

func ExampleRouteTable_DumpRouteTable() {
	ctx := context.Background()

	rtb := NewRouteTable()

	err := rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	err = rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 255),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	err = rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	err = rtb.AddRoute(ctx, &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	dumped := rtb.DumpRouteTable(ctx)
	fmt.Println(dumped)

	// Output:
	// 2001:db8::ff/128	2001:db8::ff	ifb0	1
	// 2001:db8::/32	2001:db8::1	ifb0	1
	// 192.0.2.255/32	192.0.2.255	ifb0	1
	// 192.0.2.0/24	192.0.2.1	ifb0	1
}

func ExampleRouteTable_ClearRoutes() {
	ctx := context.Background()

	rtb := NewRouteTable()

	dst1 := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	err := rtb.AddRoute(ctx, &Route{
		Destination:      dst1,
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	dst2 := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 255),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}
	err = rtb.AddRoute(ctx, &Route{
		Destination:      dst2,
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	matched, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.Unwrap())

	matched, err = rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.Unwrap())

	rtb.ClearRoutes(ctx)

	matched, err = rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.IsSome())

	matched, err = rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.IsSome())

	// Output:
	// 192.0.2.0/24	192.0.2.1	ifb0	1
	// 192.0.2.255/32	192.0.2.255	ifb0	1
	// false
	// false
}

func ExampleRoute_MarshalJSON() {
	r := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0).To4(),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}

	marshalled, err := json.Marshal(r)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", marshalled)

	// Output:
	// {"destination":"192.0.2.0/24","gateway":"192.0.2.1","networkInterface":"ifb0","metric":1}
}

func ExampleRoute_UnmarshalJSON() {
	var r Route
	err := json.Unmarshal([]byte(`{"destination":"192.0.2.0/24","gateway":"192.0.2.1","networkInterface":"ifb0","metric":1}`), &r)
	if err != nil {
		panic(err)
	}

	fmt.Printf("destination: %s\n", r.Destination.String())
	fmt.Printf("gateway: %s\n", r.Gateway.String())
	fmt.Printf("networkInterface: %s\n", r.NetworkInterface)
	fmt.Printf("metric: %d\n", r.Metric)

	// Output:
	// destination: 192.0.2.0/24
	// gateway: 192.0.2.1
	// networkInterface: ifb0
	// metric: 1
}
