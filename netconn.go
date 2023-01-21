package mockconn

import (
	"errors"
	"fmt"
	"net"
	"time"
)

var (
	ErrConnNotEstablished error = errors.New("NetConn is not established")
)

type NetConn struct {
	sendConn *UniConn
	recvConn *UniConn
}

// An implement of net.Conn interface
func NewNetConn(sendConn, recvConn *UniConn) *NetConn {
	nc := &NetConn{sendConn: sendConn, recvConn: recvConn}
	return nc
}

func (nc *NetConn) Write(b []byte) (n int, err error) {
	if nc.sendConn == nil {
		return 0, ErrConnNotEstablished
	}
	return nc.sendConn.Write(b)
}

func (nc *NetConn) Read(b []byte) (n int, err error) {
	if nc.recvConn == nil {
		return 0, ErrConnNotEstablished
	}
	return nc.recvConn.Read(b)
}

func (nc *NetConn) Close() error {
	if nc.sendConn == nil || nc.recvConn == nil {
		return ErrConnNotEstablished
	}

	nc.sendConn.CloseWrite()
	nc.recvConn.CloseRead()

	return nil
}

func (nc *NetConn) LocalAddr() net.Addr {
	if nc.sendConn == nil {
		return nil
	}
	return nc.sendConn.LocalAddr()
}

func (nc *NetConn) RemoteAddr() net.Addr {
	if nc.sendConn == nil {
		return nil
	}
	return nc.sendConn.RemoteAddr()
}

func (nc *NetConn) SetDeadline(t time.Time) error {
	if nc.sendConn == nil {
		return ErrConnNotEstablished
	}
	if nc.recvConn == nil {
		return ErrConnNotEstablished
	}

	err := nc.sendConn.SetDeadline(t)
	if err != nil {
		return err
	}
	err = nc.recvConn.SetDeadline(t)
	if err != nil {
		return err
	}

	return nil
}

func (nc *NetConn) SetReadDeadline(t time.Time) error {
	if nc.recvConn == nil {
		return ErrConnNotEstablished
	}
	return nc.recvConn.SetReadDeadline(t)
}

func (nc *NetConn) SetWriteDeadline(t time.Time) error {
	if nc.sendConn == nil {
		return ErrConnNotEstablished
	}
	return nc.sendConn.SetWriteDeadline(t)
}

func (nc *NetConn) PrintMetrics() {
	if nc.recvConn == nil {
		return
	}
	nc.recvConn.PrintMetrics()
}

func (nc *NetConn) String() string {
	return fmt.Sprintf("NetConn endpoint %v", nc.LocalAddr())
}
