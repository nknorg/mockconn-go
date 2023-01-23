package mockconn

import (
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

var (
	ErrConnNotEstablished error = errors.New("NetConn is not established")
)

type NetConn struct {
	sendConn *UniConn
	recvConn *UniConn

	readMu     sync.RWMutex
	pauseRead  bool
	writeMu    sync.RWMutex
	pauseWrite bool
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

	nc.writeMu.RLock()
	defer nc.writeMu.RUnlock()
	if nc.pauseWrite {
		return 0, errors.New("NetConn writing is paused")
	}

	return nc.sendConn.Write(b)
}

func (nc *NetConn) Read(b []byte) (n int, err error) {
	if nc.recvConn == nil {
		return 0, ErrConnNotEstablished
	}

	nc.readMu.RLock()
	defer nc.readMu.RUnlock()
	if nc.pauseRead {
		return 0, errors.New("NetConn reading is paused")
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

func (nc *NetConn) CloseRead() error {
	return nc.recvConn.CloseRead()
}

func (nc *NetConn) CloseWrite() error {
	return nc.sendConn.CloseWrite()
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

// The following functions are to better stimulating of network exceptions
func (nc *NetConn) PauseRead() {
	nc.readMu.Lock()
	defer nc.readMu.Unlock()
	nc.pauseRead = true
}

func (nc *NetConn) ResumeRead() {
	nc.readMu.Lock()
	defer nc.readMu.Unlock()
	nc.pauseRead = false
}

func (nc *NetConn) PauseWrite() {
	nc.writeMu.Lock()
	defer nc.writeMu.Unlock()
	nc.pauseWrite = true
}

func (nc *NetConn) ResumeWrite() {
	nc.writeMu.Lock()
	defer nc.writeMu.Unlock()
	nc.pauseWrite = false
}
