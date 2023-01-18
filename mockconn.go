package mockconn

import (
	"net"
	"time"
)

// The config to mock out an connection
// A connection has two address represent two end points Addr1 and Addr2.
// Here we refer them as Addr1 and Addr2. They can be any string you would like.
type ConnConfig struct {
	Addr1      string
	Addr2      string
	Throughput uint
	BufferSize uint
	Latency    time.Duration
	Loss       float32 // 0.01 = 1%
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
