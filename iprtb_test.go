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

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	assert.NoError(t, err)

	err = rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 255),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}, net.IPv4(192, 0, 2, 255), "ifb0", 1)
	assert.NoError(t, err)

	err = rtb.AddRoute(net.IPNet{
		IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, "ifb0", 1)
	assert.NoError(t, err)

	err = rtb.AddRoute(net.IPNet{
		IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff}, "ifb0", 1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 254))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		// longest match
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 255))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 255), maybeMatchedRoute.Unwrap().Gateway)
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
		assert.Equal(t, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xfe})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}, maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff})
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff}, maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_RemoveRoute(t *testing.T) {
	rtb := NewRouteTable()

	dst1 := net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	dst2 := net.IPNet{
		IP:   net.IPv4(192, 0, 3, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}

	err := rtb.AddRoute(dst1, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	assert.NoError(t, err)

	err = rtb.AddRoute(dst2, net.IPv4(192, 0, 3, 1), "ifb0", 1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 3, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(dst1)
	assert.NoError(t, err)
	err = rtb.RemoveRoute(dst2)
	assert.NoError(t, err)
	notExistingRoute := net.IPNet{
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

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	assert.NoError(t, err)

	err = rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.IPv4Mask(0, 0, 0, 0),
	}, net.IPv4(0, 0, 0, 0), "ifb0", 1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(0, 0, 0, 0), maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_RemoveRoute_DefaultRoute(t *testing.T) {
	rtb := NewRouteTable()

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.IPv4Mask(0, 0, 0, 0),
	}, net.IPv4(0, 0, 0, 0), "ifb0", 1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 3, 0))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(0, 0, 0, 0), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRoute(net.IPNet{
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

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	}, net.IP{0xff, 0xff, 0xff, 0xff, 0xff}, "ifb0", 1)
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_MatchRoute_WithInvalidIPv6(t *testing.T) {
	rtb := NewRouteTable()

	_, err := rtb.MatchRoute(net.IP{0xff, 0xff, 0xff, 0xff, 0xff})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_RemoveRoute_WithInvalidIPv6(t *testing.T) {
	rtb := NewRouteTable()

	err := rtb.RemoveRoute(net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	})
	assert.ErrorIs(t, err, ErrInvalidIPv6Length)
}

func TestRouteTable_AddRoute_ForUpdate(t *testing.T) {
	rtb := NewRouteTable()

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 2), "ifb0", 1)
	assert.NoError(t, err)

	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 2), maybeMatchedRoute.Unwrap().Gateway)
	}
}
