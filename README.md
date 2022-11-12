# go-iprtb [![.github/workflows/check.yml](https://github.com/moznion/go-iprtb/actions/workflows/check.yml/badge.svg)](https://github.com/moznion/go-iprtb/actions/workflows/check.yml) [![codecov](https://codecov.io/gh/moznion/go-iprtb/branch/main/graph/badge.svg?token=S3UWM0Y3LF)](https://codecov.io/gh/moznion/go-iprtb) [![GoDoc](https://godoc.org/github.com/moznion/go-iprtb?status.svg)](https://godoc.org/github.com/moznion/go-iprtb)

Pure go implementation of the IP routing table. This implementation uses a prefix tree as a backend data structure to find/match the routes.

NOTE: This library is isolated from the OS-level routing table. This intends to provide the routing table function on the user-level code.

## Synopsis

```go
import (
	"fmt"
	"net"

	"github.com/moznion/go-iprtb"
)

func main() {
	rtb := iprtb.NewRouteTable()

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
	fmt.Println(maybeRoute.IsSome()) // => true
	fmt.Println(maybeRoute.Unwrap().String()) // => 192.0.2.0/24	192.0.2.1	ifb0	1

	maybeRoute, err = rtb.MatchRoute(net.IPv4(192, 0, 2, 254))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome()) // => true
	fmt.Println(maybeRoute.Unwrap().String()) // => 192.0.2.0/24	192.0.2.1	ifb0	1

	// longest match
	maybeRoute, err = rtb.MatchRoute(net.IPv4(192, 0, 2, 255))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome()) // => true
	fmt.Println(maybeRoute.Unwrap().String()) // => 192.0.2.255/32	192.0.2.255	ifb0	1

	// not routes
	maybeRoute, err = rtb.MatchRoute(net.IPv4(198, 51, 100, 123))
	if err != nil {
		panic(err)
	}
	fmt.Println(maybeRoute.IsSome()) // => false
}
```

And please see also [examples_test.go](./examples_test.go).

## Docs

[![GoDoc](https://godoc.org/github.com/moznion/go-iprtb?status.svg)](https://godoc.org/github.com/moznion/go-iprtb)

## Author

moznion (<moznion@mail.moznion.net>)

