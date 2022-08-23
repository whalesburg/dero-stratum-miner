//go:build linux && (arm || arm64)
// +build linux
// +build arm arm64

package dns

import (
	"context"
	"fmt"
	"net"
)

func BootstrapDNS(ip string) bool {
	var dialer net.Dialer
	net.DefaultResolver = &net.Resolver{
		PreferGo: false,
		Dial: func(context context.Context, _, _ string) (net.Conn, error) {
			conn, err := dialer.DialContext(context, "udp", fmt.Sprintf("%s:53", ip))
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
	}
	return true
}
