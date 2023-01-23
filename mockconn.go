package mockconn

import (
	"net"
	"time"
)

// The config to mock out an connection
// A connection has two address represent two endpoints Addr1 and Addr2.
// Here we refer them as Addr1 and Addr2. They can be any string you would like.
type ConnConfig struct {
	Addr1        string        // endpoint 1 address
	Addr2        string        // endpoint 2 address
	Throughput   uint          // throughput by packets/second
	BufferSize   uint          // BufferSize used int connection. If it is not set, a default value will be computed.
	Latency      time.Duration // Latency is the duration which the packet travels from endpoint 1 to endpoint 2.
	Loss         float32       // loss rate, 0.01 = 1%
	WriteTimeout time.Duration // set default timeout for writing, without timeout if zero
	ReadTimeout  time.Duration // set default timeout for reading, without timeout if zero
}

// Mock network connection
// Return two net.Conn(s) which represent two endpoints of this connection.
func NewMockConn(conf *ConnConfig) (net.Conn, net.Conn, error) {
	l2r, err := NewUniConn(conf)
	if err != nil {
		return nil, nil, err
	}

	conf2 := *conf
	conf2.Addr1, conf2.Addr2 = conf2.Addr2, conf2.Addr1 // switch address
	r2l, err := NewUniConn(&conf2)
	if err != nil {
		return nil, nil, err
	}

	localEndpoint := NewNetConn(l2r, r2l)
	remoteEndpoint := NewNetConn(r2l, l2r)

	return localEndpoint, remoteEndpoint, nil
}
