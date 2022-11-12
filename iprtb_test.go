package iprtb

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouteTable_MatchRoute(t *testing.T) {
	rtb := NewRouteTable()

	rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 0),
		Mask: net.IPv4Mask(255, 255, 255, 0),
	}, net.IPv4(192, 0, 2, 1), "ifb0", 1)
	rtb.AddRoute(net.IPNet{
		IP:   net.IPv4(192, 0, 2, 200),
		Mask: net.IPv4Mask(255, 255, 255, 255),
	}, net.IPv4(192, 0, 2, 200), "ifb0", 1)

	{
		maybeMatchedRoute := rtb.MatchRoute(net.IPv4(192, 0, 2, 100))
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 1), maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute := rtb.MatchRoute(net.IPv4(192, 0, 2, 200))
		assert.True(t, maybeMatchedRoute.IsSome())
		assert.Equal(t, net.IPv4(192, 0, 2, 200), maybeMatchedRoute.Unwrap().Gateway)
	}

	{
		maybeMatchedRoute := rtb.MatchRoute(net.IPv4(198, 51, 100, 100))
		assert.True(t, maybeMatchedRoute.IsNone())
	}
}
