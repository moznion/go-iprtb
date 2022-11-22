package iprtb

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteTable_MatchRoute(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
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
	err := rtb.AddRoute(ctx, route1)
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
	err = rtb.AddRoute(ctx, route2)
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
	err = rtb.AddRoute(ctx, route3)
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
	err = rtb.AddRoute(ctx, route4)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 254))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		// longest match
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 255))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(198, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route3.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route3.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route4.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_RemoveRoute(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route1)
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
	err = rtb.AddRoute(ctx, route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(ctx, route1.Destination)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(ctx, route2.Destination)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}

	notExistingRoute := &net.IPNet{
		IP:   net.IPv4(192, 0, 3, 255),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}
	err = rtb.RemoveRoute(ctx, notExistingRoute)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
}

func TestRouteTable_MatchRoute_DefaultRoute(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route1)
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
	err = rtb.AddRoute(ctx, route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_RemoveRoute_DefaultRoute(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(ctx, &net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.IPv4Mask(0, 0, 0, 0),
	})
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
}

func TestRouteTable_AddRoute_WithInvalidIPv6(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_MatchRoute_WithInvalidIPv6(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	_, err := rtb.MatchRoute(ctx, net.IP{0xff, 0xff, 0xff, 0xff, 0xff})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_RemoveRoute_WithInvalidIPv6(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	err := rtb.RemoveRoute(ctx, &net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_AddRoute_ForUpdate(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
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
	err = rtb.AddRoute(ctx, route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_FindRoute(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()
	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.False(t, found)
	}
	{
		found, err := rtb.FindRoute(ctx, net.IPv4(10, 0, 2, 100))
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
	err := rtb.AddRoute(ctx, route1)
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
	err = rtb.AddRoute(ctx, route2)
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
	err = rtb.AddRoute(ctx, route3)
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
	err = rtb.AddRoute(ctx, route4)
	assert.NoError(t, err)

	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 254))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		// longest match (but this should be early break)
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 255))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.False(t, found)
	}

	{
		found, err := rtb.FindRoute(ctx, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02})
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(ctx, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe})
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(ctx, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff})
		assert.NoError(t, err)
		assert.True(t, found)
	}
}

func TestRouteTable_FindRoute_DefaultRoute(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route)
	assert.NoError(t, err)

	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, found)
	}

	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, found)
	}
}

func TestRouteTable_FindRoute_ExactMatch(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRoute(ctx, route)
	assert.NoError(t, err)

	{
		found, err := rtb.FindRoute(ctx, net.IPv4(192, 0, 2, 123))
		assert.NoError(t, err)
		assert.True(t, found)
	}
}

func TestRouteTable_FindRoute_WithInvalidIpv6(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	_, err := rtb.FindRoute(ctx, net.IP{0xff, 0xff, 0xff, 0xff, 0xff})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_WithLabel(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRouteWithLabel(ctx, label, route)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
		assert.NotEmpty(t, rtb.label2Destination)
		assert.NotEmpty(t, rtb.destination2Label)
	}

	err = rtb.UpdateRouteByLabel(ctx, label, net.IPv4(192, 0, 2, 2), "ifb0", 1)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 2), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRouteByLabel(ctx, label)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.False(t, maybeMatchedRoute.IsSome())
		assert.Empty(t, rtb.label2Destination)
		assert.Empty(t, rtb.destination2Label)
	}
}

func TestRouteTable_WithNotExistedLabel(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRouteWithLabel(ctx, label, route)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.UpdateRouteByLabel(ctx, "__invalid_label__", net.IPv4(192, 0, 2, 2), "ifb0", 1)
	assert.NoError(t, err)
	{
		// should not be affected
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRouteByLabel(ctx, "__invalid_label__")
	assert.NoError(t, err)
	{
		// should not be affected
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_AddRouteWithLabel_WithInvalidIpv6(t *testing.T) {
	ctx := context.Background()

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
	err := rtb.AddRouteWithLabel(ctx, "__label__", route)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_RemoveRouteByLabel_WithInvalidIpv6(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	label := "__label__"
	rtb.label2Destination[label] = &net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	}

	err := rtb.RemoveRouteByLabel(ctx, label)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_DumpRouteTable(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()
	dumped := rtb.DumpRouteTable(ctx)
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
	err := rtb.AddRoute(ctx, route1)
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
	err = rtb.AddRoute(ctx, route2)
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
	err = rtb.AddRoute(ctx, route3)
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
	err = rtb.AddRoute(ctx, route4)
	assert.NoError(t, err)

	dumped = rtb.DumpRouteTable(ctx)
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
	ctx := context.Background()

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
	err := rtb.AddRouteWithLabel(ctx, "__label1__", route1)
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
	err = rtb.AddRouteWithLabel(ctx, "__label2__", route2)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route1.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, route2.Gateway, maybeMatchedRoute.Unwrap().Gateway)
	}
	assert.Len(t, rtb.label2Destination, 2)
	assert.Len(t, rtb.destination2Label, 2)

	rtb.ClearRoutes(ctx)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(ctx, net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsNone())
	}
	assert.Empty(t, rtb.label2Destination)
	assert.Empty(t, rtb.destination2Label)
}

func TestRouteTable_RemoveRoute_DestroysLabelMapping(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	dst1 := &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	err := rtb.AddRouteWithLabel(ctx, "__label1__", &Route{
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
	err = rtb.AddRouteWithLabel(ctx, "__label2__", &Route{
		Destination:      dst2,
		Gateway:          net.IPv4(192, 0, 2, 255),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	assert.NoError(t, err)

	assert.Len(t, rtb.label2Destination, 2)
	assert.Len(t, rtb.destination2Label, 2)

	err = rtb.RemoveRoute(ctx, dst1)
	assert.NoError(t, err)

	// should remove an internal mapping for a label
	assert.Len(t, rtb.label2Destination, 1)
	assert.Len(t, rtb.destination2Label, 1)

	err = rtb.RemoveRoute(ctx, dst2)
	assert.NoError(t, err)

	// should remove an internal mapping for a label (i.e. removes all)
	assert.Empty(t, rtb.label2Destination)
	assert.Empty(t, rtb.destination2Label)
}

func TestRouteTable_RemoveRoute_DoNotDeleteWhenTheTargetDoesNotTerminate(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	err := rtb.AddRouteWithLabel(ctx, "__label1__", &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb0",
		Metric:           1,
	})
	assert.NoError(t, err)

	found, _ := rtb.FindRoute(ctx, net.IP{192, 0, 2, 100})
	assert.True(t, found)

	err = rtb.RemoveRoute(ctx, &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 254, 0),
	})
	assert.NoError(t, err)

	found, _ = rtb.FindRoute(ctx, net.IP{192, 0, 2, 100})
	assert.True(t, found)
}

func TestRouteTable_RemoveRoute_WithNotTerminatedBranchPruning(t *testing.T) {
	ctx := context.Background()

	rtb := NewRouteTable()

	/*
	 * Routes:
	 *   - GW1: 192.0.2.0/24   => 11000000 00000000   0000001[0] 00000000
	 *   - GW2: 192.0.0.0/22   => 11000000 00000000   00000[0]00 00000000
	 *   - GW3: 192.0.0.0/16   => 11000000 0000000[0] 00000000   00000000
	 *   - GW4: 192.0.128.0/17 => 11000000 00000000   [1]0000000 00000000
	 *
	 * Diagram:
	 *                            R
	 *                             \
	 *                              1
	 *                               \
	 *                                1
	 *                               /
	 *                              0
	 *                             /
	 *                            0
	 *                           /
	 *                          0
	 *                         /
	 *                        0
	 *                       /
	 *                      0
	 *                     /
	 *                    0
	 *                   /
	 *                  0... (7 nodes)
	 *                 /
	 *               [0] GW3 <= 3) if this route has been removed, this node should be keeped but the route terminal has to be removed
	 *               / \
	 *              0  [1] GW4 <= 4) if this route has been removed, the above nodes until the root node also should be removed
	 *             /
	 *            0
	 *           /
	 *          0
	 *         /
	 *        0
	 *       /
	 *      0
	 *     /
	 *   [0] GW2 <= 2) if this route has been removed, the above "0" nodes until "GW3" node also should be removed
	 *     \
	 *      1 <= † not terminal node
	 *     /
	 *   [0] GW1 <= 1) if this route has been removed, the above "1" node also should be removed
	 */

	label1 := "label1"
	err := rtb.AddRouteWithLabel(ctx, label1, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 2, 0),
			Mask: net.IPv4Mask(255, 255, 255, 0),
		},
		Gateway:          net.IPv4(192, 0, 2, 1),
		NetworkInterface: "ifb1",
		Metric:           1,
	})
	assert.NoError(t, err)

	label2 := "label2"
	err = rtb.AddRouteWithLabel(ctx, label2, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 0, 0),
			Mask: net.IPv4Mask(255, 255, 252, 0),
		},
		Gateway:          net.IPv4(192, 0, 0, 0),
		NetworkInterface: "ifb2",
		Metric:           1,
	})
	assert.NoError(t, err)

	label3 := "label3"
	err = rtb.AddRouteWithLabel(ctx, label3, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 0, 0),
			Mask: net.IPv4Mask(255, 255, 0, 0),
		},
		Gateway:          net.IPv4(192, 255, 0, 0),
		NetworkInterface: "ifb3",
		Metric:           1,
	})
	assert.NoError(t, err)

	label4 := "label4"
	err = rtb.AddRouteWithLabel(ctx, label4, &Route{
		Destination: &net.IPNet{
			IP:   net.IPv4(192, 0, 128, 0),
			Mask: net.IPv4Mask(255, 255, 128, 0),
		},
		Gateway:          net.IPv4(192, 255, 0, 0),
		NetworkInterface: "ifb4",
		Metric:           1,
	})
	assert.NoError(t, err)

	// attempt to remove by not terminated node, but it should do nothing
	err = rtb.RemoveRoute(ctx, &net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 254, 0),
	})
	assert.NoError(t, err)
	found, _ := rtb.FindRoute(ctx, net.IP{192, 0, 2, 100})
	assert.True(t, found)

	n := rtb.routes.oneBitNode.oneBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
		zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
		zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.oneBitNode.zeroBitNode
	assert.NotNil(t, n)
	assert.NotNil(t, n.route)
	assert.Nil(t, n.zeroBitNode)
	assert.Nil(t, n.oneBitNode)
	maybeMatched, _ := rtb.MatchRoute(ctx, net.IP{192, 0, 2, 100})
	assert.EqualValues(t, "ifb1", maybeMatched.Unwrap().NetworkInterface)

	err = rtb.RemoveRouteByLabel(ctx, label1)
	assert.NoError(t, err)
	assert.Nil(t,
		rtb.routes.oneBitNode.oneBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
			zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
			zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.oneBitNode,
		"the not terminal branch should be removed",
	)
	n = rtb.routes.oneBitNode.oneBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
		zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
		zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode
	assert.NotNil(t, n)
	assert.NotNil(t, n.route)
	assert.Nil(t, n.zeroBitNode)
	assert.Nil(t, n.oneBitNode)
	maybeMatched, _ = rtb.MatchRoute(ctx, net.IP{192, 0, 2, 100})
	assert.EqualValues(t, "ifb2", maybeMatched.Unwrap().NetworkInterface)

	err = rtb.RemoveRouteByLabel(ctx, label2)
	assert.Nil(t,
		rtb.routes.oneBitNode.oneBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
			zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
			zeroBitNode,
		"the not terminal branch should be removed",
	)
	assert.NoError(t, err)
	n = rtb.routes.oneBitNode.oneBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
		zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode
	assert.NotNil(t, n)
	assert.NotNil(t, n.route)
	assert.Nil(t, n.zeroBitNode)
	assert.NotNil(t, n.oneBitNode)
	maybeMatched, _ = rtb.MatchRoute(ctx, net.IP{192, 0, 2, 100})
	assert.EqualValues(t, "ifb3", maybeMatched.Unwrap().NetworkInterface)

	err = rtb.RemoveRouteByLabel(ctx, label3)
	assert.NoError(t, err)
	n = rtb.routes.oneBitNode.oneBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.
		zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode.zeroBitNode
	assert.NotNil(t, n)
	assert.Nil(t, n.route)
	assert.Nil(t, n.zeroBitNode)
	assert.NotNil(t, n.oneBitNode)
	found, _ = rtb.FindRoute(ctx, net.IP{192, 0, 2, 100})
	assert.False(t, found)
	maybeMatched, _ = rtb.MatchRoute(ctx, net.IP{192, 0, 128, 1})
	assert.EqualValues(t, "ifb4", maybeMatched.Unwrap().NetworkInterface)

	err = rtb.RemoveRouteByLabel(ctx, label4)
	assert.NoError(t, err)
	n = rtb.routes // root: there is no child route under the root node, so all nodes should be removed
	assert.NotNil(t, n)
	assert.Nil(t, n.route)
	assert.Nil(t, n.zeroBitNode)
	assert.Nil(t, n.oneBitNode)
	found, _ = rtb.FindRoute(ctx, net.IP{192, 0, 128, 1})
	assert.False(t, found)
}
