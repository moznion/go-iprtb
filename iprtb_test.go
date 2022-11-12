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

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(0, 0, 0, 0),
		Mask: net.IPv4Mask(0, 0, 0, 0),
	}, net.IPv4(0, 0, 0, 0), "ifb0", 1)
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

	err := rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 123),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}, net.IPv4(192, 0, 2, 123), "ifb0", 1)
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
	err := rtb.AddRouteWithLabel(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1, label)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
		assert.NotEmpty(t, rtb.label2Destination)
	}

	err = rtb.UpdateByLabel(label, net.IPv4(192, 0, 2, 2), "ifb0", 1)
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
	}
}

func TestRouteTable_WithNotExistedLabel(t *testing.T) {
	label := "label-1"

	rtb := NewRouteTable()
	err := rtb.AddRouteWithLabel(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1, label)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.UpdateByLabel("__invalid_label__", net.IPv4(192, 0, 2, 2), "ifb0", 1)
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	err = rtb.RemoveRouteByLabel("__invalid_label__")
	assert.NoError(t, err)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}
}

func TestRouteTable_AddRouteWithLabel_WithInvalidIpv6(t *testing.T) {
	rtb := NewRouteTable()
	err := rtb.AddRouteWithLabel(net.IPNet{
		IP:   net.IP{0xff, 0xff, 0xff, 0xff, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff},
	}, net.IP{0xff, 0xff, 0xff, 0xff, 0xff}, "ifb0", 1, "__label__")
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

	nwInterface := "ifb0"
	metric := 1

	dst1 := net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}
	gw1 := net.IPv4(192, 0, 2, 1)
	err := rtb.AddRoute(dst1, gw1, nwInterface, metric)
	assert.NoError(t, err)

	dst2 := net.IPNet{
		IP:   net.IPv4(192, 0, 2, 255),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}
	gw2 := net.IPv4(192, 0, 2, 255)
	err = rtb.AddRoute(dst2, gw2, nwInterface, metric)
	assert.NoError(t, err)

	dst3 := net.IPNet{
		IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
	}
	gw3 := net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x01}
	err = rtb.AddRoute(dst3, gw3, nwInterface, metric)
	assert.NoError(t, err)

	dst4 := net.IPNet{
		IP:   net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff},
		Mask: net.IPMask{0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
	}
	gw4 := net.IP{0x20, 0x01, 0x0d, 0xb8, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xff}
	err = rtb.AddRoute(dst4, gw4, nwInterface, metric)
	assert.NoError(t, err)

	dumped = rtb.DumpRouteTable()
	assert.Len(t, dumped, 4)
	assert.Contains(t, dumped, &RouteEntry{
		Destination: dst1,
		Gateway:     gw1,
		NwInterface: nwInterface,
		Metric:      metric,
	})
	assert.Contains(t, dumped, &RouteEntry{
		Destination: dst2,
		Gateway:     gw2,
		NwInterface: nwInterface,
		Metric:      metric,
	})
	assert.Contains(t, dumped, &RouteEntry{
		Destination: dst3,
		Gateway:     gw3,
		NwInterface: nwInterface,
		Metric:      metric,
	})
	assert.Contains(t, dumped, &RouteEntry{
		Destination: dst4,
		Gateway:     gw4,
		NwInterface: nwInterface,
		Metric:      metric,
	})
}
