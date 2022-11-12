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
		assert.False(t, maybeMatchedRoute.IsSome())
	}

	rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 255),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}, net.IPv4(192, 0, 2, 255), "ifb0", 1)

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
}

func TestRouteTable_RemoveRoute(t *testing.T) {
	rtb := NewRouteTable()

	dst := net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}

	rtb.AddRoute(dst, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	rtb.RemoveRoute(dst)
	{
		maybeMatchedRoute, err := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.NoError(t, err)
		assert.False(t, maybeMatchedRoute.IsSome())
	}
}
