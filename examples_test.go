package iprtb

import (
	"fmt"
	"net"
)

func ExampleRouteTable_MatchRoute() {
	rtb := NewRouteTable()

	err := rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	err = rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	maybeRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())
	fmt.Println(maybeRoute.Unwrap().String())

	maybeRoute, err = rtb.MatchRoute(net.IPv4(192, 0, 2, 254))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())
	fmt.Println(maybeRoute.Unwrap().String())

	// longest match
	maybeRoute, err = rtb.MatchRoute(net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome())
	fmt.Println(maybeRoute.Unwrap().String())

	// not routes
	maybeRoute, err = rtb.MatchRoute(net.IPv4(198, 51, 100, 123))
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
	rtb := NewRouteTable()

	err := rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	err = rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	found, err = rtb.FindRoute(net.IPv4(192, 0, 2, 254))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// longest match (but this is early broke)
	found, err = rtb.FindRoute(net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// not routes
	found, err = rtb.FindRoute(net.IPv4(198, 51, 100, 123))
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
	rtb := NewRouteTable()

	dst := net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	err := rtb.AddRoute(&Route{
		Destination:      dst,
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	if err != nil {
		panic(err)
	}

	found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	err = rtb.RemoveRoute(dst)
	if err != nil {
		panic(err)
	}

	// route has been removed
	found, err = rtb.FindRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// Output:
	// true
	// false
}

func ExampleRouteTable_RemoveRouteByLabel() {
	rtb := NewRouteTable()

	label := "__label__"
	err := rtb.AddRouteWithLabel(label, &Route{
		Destination: net.IPNet{
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

	found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	err = rtb.RemoveRouteByLabel(label)
	if err != nil {
		panic(err)
	}

	// route has been removed
	found, err = rtb.FindRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(found)

	// Output:
	// true
	// false
}

func ExampleRouteTable_UpdateRouteByLabel() {
	rtb := NewRouteTable()

	label := "__label__"
	err := rtb.AddRouteWithLabel(label, &Route{
		Destination: net.IPNet{
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

	matched, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.Unwrap().String())

	err = rtb.UpdateRouteByLabel(label, net.IPv4(192, 0, 2, 2), "ifb1", 2)
	if err != nil {
		panic(err)
	}
	matched, err = rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
	if err != nil {
		panic(err)
	}
	fmt.Println(matched.Unwrap().String())

	// Output:
	// 192.0.2.0/24	192.0.2.1	ifb0	1
	// 192.0.2.0/24	192.0.2.2	ifb1	2
}

func ExampleRouteTable_DumpRouteTable() {
	rtb := NewRouteTable()

	err := rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	err = rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	err = rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	err = rtb.AddRoute(&Route{
		Destination: net.IPNet{
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

	dumped := rtb.DumpRouteTable()
	fmt.Println(dumped)

	// Output:
	// 2001:db8::ff/128	2001:db8::ff	ifb0	1
	// 2001:db8::/32	2001:db8::1	ifb0	1
	// 192.0.2.255/32	192.0.2.255	ifb0	1
	// 192.0.2.0/24	192.0.2.1	ifb0	1
}
