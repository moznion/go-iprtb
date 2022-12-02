package iprtb

import (
	"encoding/json"
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRoute_MarshalJSON(t *testing.T) {
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
	assert.NoError(t, err)

	var unmarshalled Route
	err = json.Unmarshal(marshalled, &unmarshalled)
	assert.NoError(t, err)
	assert.EqualValues(t, r, &unmarshalled)
}

func TestRoute_UnmarshalJSON_FailToUnmarshalRouteJson(t *testing.T) {
	var r Route
	err := json.Unmarshal([]byte(`{"metric":"__invalid-NaN__"}`), &r)
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), "failed to unmarshal Route:"))

	err = json.Unmarshal([]byte(`{"destination":"192.0.2.0/INVALID"}`), &r)
	assert.Error(t, err)
	assert.True(t, strings.HasPrefix(err.Error(), `failed to unmarshal Route; it cannot parse the value of "destination" property as net.IPNet:`))
}
