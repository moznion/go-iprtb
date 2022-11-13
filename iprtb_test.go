package iprtb

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteTable_MatchRoute(t *testing.T) {
	rtb := NewRouteTable()
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}

	route1 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route1)
	assert.NoError(t, err)

	route2 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 255),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route2)
	assert.NoError(t, err)

	route3 := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route3)
	assert.NoError(t, err)

	route4 := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route4)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 254))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		// longest match
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 255))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(198, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route3.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route3.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route4.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_RemoveRoute(t *testing.T) {
	rtb := NewRouteTable()

	route1 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route1)
	assert.NoError(t, err)

	route2 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 3, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 3, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(route1.Destination)
	assert.NoError(t, err)
	err = rtb.RemoveRoute(route2.Destination)
	assert.NoError(t, err)

	notExistingRoute := &net.IPNet{
		IP:   net.IPv4(192, 0, 3, 255),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}
	err = rtb.RemoveRoute(notExistingRoute)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
}

func TestRouteTable_MatchRoute_DefaultRoute(t *testing.T) {
	rtb := NewRouteTable()

	route1 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route1)
	assert.NoError(t, err)

	route2 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(0, 0, 0, 0),
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Gateway:          net.IPv4(0, 0, 0, 0),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_RemoveRoute_DefaultRoute(t *testing.T) {
	rtb := NewRouteTable()

	var route = &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(0, 0, 0, 0),
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Gateway:          net.IPv4(0, 0, 0, 0),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(&net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.IPv4Mask(0, 0, 0, 0),
	})
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
}

func TestRouteTable_AddRoute_WithInvalidIPv6(t *testing.T) {
	rtb := NewRouteTable()

	route := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
		},
		Gateway:          net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_MatchRoute_WithInvalidIPv6(t *testing.T) {
	rtb := NewRouteTable()

	_, err := rtb.MatchRoute(net.IP{0xff, 0xff, 0xff, 0xff, 0xff})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_RemoveRoute_WithInvalidIPv6(t *testing.T) {
	rtb := NewRouteTable()

	err := rtb.RemoveRoute(&net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_AddRoute_ForUpdate(t *testing.T) {
	rtb := NewRouteTable()

	dst := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}

	route1 := &Route{
		Destination:      dst,
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	route2 := &Route{
		Destination:      dst,
		Gateway:          net.IPv4(192, 0, 2, 2),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_FindRoute(t *testing.T) {
	rtb := NewRouteTable()
	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.False(t, found)
	}
	{
		found, err := rtb.FindRoute(net.IPv4(10, 0, 2, 100))
		assert.NoError(t, err)
		assert.False(t, found)
	}

	route1 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route1)
	assert.NoError(t, err)

	route2 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 255),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route2)
	assert.NoError(t, err)

	route3 := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route3)
	assert.NoError(t, err)

	route4 := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRoute(route4)
	assert.NoError(t, err)

	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 254))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		// longest match (but this should be early break)
		found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 255))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.False(t, found)
	}

	{
		found, err := rtb.FindRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe})
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff})
		assert.NoError(t, err)
		assert.True(t, found)
	}
}

func TestRouteTable_FindRoute_DefaultRoute(t *testing.T) {
	rtb := NewRouteTable()

	route := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(0, 0, 0, 0),
			Mask: net.IPv4Mask(0, 0, 0, 0),
		},
		Gateway:          net.IPv4(0, 0, 0, 0),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route)
	assert.NoError(t, err)

	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, found)
	}
}

func TestRouteTable_FindRoute_ExactMatch(t *testing.T) {
	rtb := NewRouteTable()

	route := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 123),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 123),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRoute(route)
	assert.NoError(t, err)

	{
		found, err := rtb.FindRoute(net.IPv4(192, 0, 2, 123))
		assert.NoError(t, err)
		assert.True(t, found)
	}
}

func TestRouteTable_FindRoute_WithInvalidIpv6(t *testing.T) {
	rtb := NewRouteTable()

	_, err := rtb.FindRoute(net.IP{0xff, 0xff, 0xff, 0xff, 0xff})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_WithLabel(t *testing.T) {
	label := "label-1"

	rtb := NewRouteTable()

	route := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRouteWithLabel(label, route)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
		assert.NotEmpty(t, rtb.label2Destination)
		assert.NotEmpty(t, rtb.destination2Label)
	}

	err = rtb.UpdateRouteByLabel(label, net.IPv4(192, 0, 2, 2), "ifb0", 1)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 2), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRouteByLabel(label)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.False(t, maybeMatchedRoute.IsSome())
		assert.Empty(t, rtb.label2Destination)
		assert.Empty(t, rtb.destination2Label)
	}
}

func TestRouteTable_WithNotExistedLabel(t *testing.T) {
	label := "label-1"

	rtb := NewRouteTable()

	route := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRouteWithLabel(label, route)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.UpdateRouteByLabel("__invalid_label__", net.IPv4(192, 0, 2, 2), "ifb0", 1)
	assert.NoError(t, err)
	{
		// should not be affected
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRouteByLabel("__invalid_label__")
	assert.NoError(t, err)
	{
		// should not be affected
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_AddRouteWithLabel_WithInvalidIpv6(t *testing.T) {
	rtb := NewRouteTable()

	route := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
		},
		Gateway:          net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRouteWithLabel("__label__", route)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_RemoveRouteByLabel_WithInvalidIpv6(t *testing.T) {
	rtb := NewRouteTable()

	label := "__label__"
	rtb.label2Destination[label] = &net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	}

	err := rtb.RemoveRouteByLabel(label)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_DumpRouteTable(t *testing.T) {
	rtb := NewRouteTable()
	dumped := rtb.DumpRouteTable()
	assert.Empty(t, dumped)
	assert.Empty(t, dumped.String())

	nwInterface := "ifb0"
	metric := 1

	route1 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: nwInterface,
		Metric:           metric,
	}
	err := rtb.AddRoute(route1)
	assert.NoError(t, err)

	route2 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 255),
			Mask: net.IPv4Mask(255, 255, 255, 255),
		},
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: nwInterface,
		Metric:           metric,
	}
	err = rtb.AddRoute(route2)
	assert.NoError(t, err)

	route3 := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01},
		NetworkInterface: nwInterface,
		Metric:           metric,
	}
	err = rtb.AddRoute(route3)
	assert.NoError(t, err)

	route4 := &Route{
		Destination: &net.IPNet{
			IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
			Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
		},
		Gateway:          net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		NetworkInterface: nwInterface,
		Metric:           metric,
	}
	err = rtb.AddRoute(route4)
	assert.NoError(t, err)

	dumped = rtb.DumpRouteTable()
	assert.Len(t, dumped, 4)
	assert.Contains(t, dumped, route1)
	assert.Contains(t, dumped, route2)
	assert.Contains(t, dumped, route3)
	assert.Contains(t, dumped, route4)
	assert.Equal(t, `2001:db8::ff/128	2001:db8::ff	ifb0	1
2001:db8::/32	2001:db8::1	ifb0	1
192.0.2.255/32	192.0.2.255	ifb0	1
192.0.2.0/24	192.0.2.1	ifb0	1
`, dumped.String())
}

func TestRouteTable_ClearRoutes(t *testing.T) {
	rtb := NewRouteTable()

	route1 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err := rtb.AddRouteWithLabel("__label1__", route1)
	assert.NoError(t, err)

	route2 := &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 3, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 3, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	}
	err = rtb.AddRouteWithLabel("__label2__", route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
	assert.Len(t, rtb.label2Destination, 2)
	assert.Len(t, rtb.destination2Label, 2)

	rtb.ClearRoutes()
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	assert.Empty(t, rtb.label2Destination)
	assert.Empty(t, rtb.destination2Label)
}

func TestRouteTable_RemoveRoute_DestroysLabelMapping(t *testing.T) {
	rtb := NewRouteTable()

	dst1 := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	err := rtb.AddRouteWithLabel("__label1__", &Route{
		Destination:      dst1,
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	assert.NoError(t, err)

	dst2 := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 255),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	err = rtb.AddRouteWithLabel("__label2__", &Route{
		Destination:      dst2,
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	assert.NoError(t, err)

	assert.Len(t, rtb.label2Destination, 2)
	assert.Len(t, rtb.destination2Label, 2)

	err = rtb.RemoveRoute(dst1)
	assert.NoError(t, err)

	// should remove an internal mapping for a label
	assert.Len(t, rtb.label2Destination, 1)
	assert.Len(t, rtb.destination2Label, 1)

	err = rtb.RemoveRoute(dst2)
	assert.NoError(t, err)

	// should remove an internal mapping for a label (i.e. removes all)
	assert.Empty(t, rtb.label2Destination)
	assert.Empty(t, rtb.destination2Label)
}
