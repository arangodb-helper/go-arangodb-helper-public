//
// DISCLAIMER
//
// Copyright 2023 ArangoDB GmbH, Cologne, Germany
//

package conn

import (
	"context"
	"net"
	"time"

	"golang.org/x/net/proxy"
)

// TransportConnWrap instructs how to wrap net.Conn connection.
type TransportConnWrap func(net.Conn) net.Conn

type customDialer struct {
	proxy.ContextDialer
	wrapper TransportConnWrap
}

// DialContext gets connection from the net.Dialer and wraps it with the custom connection.
// which was provided in NewCustomDialer.
func (d *customDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	childConnection, err := d.ContextDialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	return d.wrapper(childConnection), nil
}

// NewContextDialer returns dialer which wraps existing connection with a network connection wrapper.
func NewContextDialer(connectionWrapper TransportConnWrap) proxy.ContextDialer {
	netDialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 100 * time.Millisecond,
	}

	if connectionWrapper != nil {
		return &customDialer{
			ContextDialer: netDialer,
			wrapper:       connectionWrapper,
		}
	}

	return netDialer
}
