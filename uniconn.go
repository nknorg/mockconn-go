package mockconn

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"time"

	"golang.org/x/time/rate"
)

var (
	zeroTime      time.Time
	ErrClosedConn error = errors.New("Connection is closed")
	ErrNilPointer error = errors.New("Data pointer is nil")
	ErrZeroLengh  error = errors.New("Zero length data to write")
	ErrUnknown    error = errors.New("UniConn unknown error")
)

// To trace time consuming.
type dataWithTime struct {
	data []byte
	t    time.Time
}

// unidirectional channel, can only send data from localAddr to remoteAddr
type UniConn struct {
	localAddr  string
	remoteAddr string

	throughput uint
	bufferSize uint
	latency    time.Duration
	loss       float32

	sendCh   chan *dataWithTime
	bufferCh chan *dataWithTime
	recvCh   chan *dataWithTime

	unreadData []byte // save unread data

	// for metrics
	nSendPacket    int64         // number of packets sent
	nRecvPacket    int64         // number of packets received
	nLoss          int64         // number of packets are random lost
	averageLatency time.Duration // average latency of all packets

	// one time deadline and cancel
	readCtx     context.Context
	readCancel  context.CancelFunc
	writeCtx    context.Context
	writeCancel context.CancelFunc

	// close UniConn
	closeWriteCtx       context.Context
	closeWriteCtxCancel context.CancelFunc
	closeReadCtx        context.Context
	closeReadCtxCancel  context.CancelFunc
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewUniConn(conf *ConnConfig) (*UniConn, error) {
	bufferSize := conf.BufferSize
	if bufferSize == 0 {
		bufferSize = uint(2 * float64(conf.Throughput) * conf.Latency.Seconds())
	}

	uc := &UniConn{throughput: conf.Throughput, bufferSize: bufferSize, latency: conf.Latency, loss: conf.Loss,
		sendCh: make(chan *dataWithTime), bufferCh: make(chan *dataWithTime, conf.BufferSize),
		recvCh: make(chan *dataWithTime), localAddr: conf.Addr1, remoteAddr: conf.Addr2}

	uc.closeWriteCtx, uc.closeWriteCtxCancel = context.WithCancel(context.Background())
	uc.closeReadCtx, uc.closeReadCtxCancel = context.WithCancel(context.Background())
	uc.SetDeadline(zeroTime)

	go uc.throughputRead()
	go uc.latencyRead()

	return uc, nil
}

func (uc *UniConn) Write(b []byte) (n int, err error) {
	if err = uc.writeCtx.Err(); err != nil {
		return 0, err
	}

	if len(b) == 0 {
		return 0, ErrZeroLengh
	}

	dt := &dataWithTime{data: b}
	select {
	case <-uc.writeCtx.Done(): // for one time deadline or for one time cancel
		return 0, uc.writeCtx.Err()

	case uc.sendCh <- dt:
		uc.nSendPacket++
	}

	return len(b), nil
}

func (uc *UniConn) randomLoss() bool {
	if uc.loss > 0 {
		l := rand.Float32()
		if l < uc.loss {
			uc.nLoss++
			return true
		}
	}

	return false
}

// The routine to stimulate throughput by rate Limiter
func (uc *UniConn) throughputRead() error {
	defer close(uc.bufferCh)

	r := rate.NewLimiter(rate.Limit(uc.throughput), 1)
	for {
		err := r.Wait(uc.closeWriteCtx)
		if err != nil {
			return err
		}

		select {
		case <-uc.closeWriteCtx.Done():
			return uc.closeWriteCtx.Err()

		case dt := <-uc.sendCh:
			if dt != nil {
				if !uc.randomLoss() {
					dt.t = time.Now()
					uc.bufferCh <- dt
				}
			}
		}
	}

	return nil
}

// The routine to stimulate latency
func (uc *UniConn) latencyRead() error {
	defer close(uc.recvCh)
	for {
		select {
		case <-uc.closeReadCtx.Done():
			return uc.closeReadCtx.Err()

		case dt := <-uc.bufferCh:
			if dt != nil {
				dur := time.Since(dt.t)
				if dur < uc.latency {
					timer := time.NewTimer(uc.latency - dur)
					select {
					case <-uc.closeReadCtx.Done():
						return uc.closeReadCtx.Err()

					case <-timer.C:
					}
				}
				uc.recvCh <- dt
			}
		}
	}
}

func (uc *UniConn) Read(b []byte) (n int, err error) {
	if err = uc.readCtx.Err(); err != nil {
		return 0, err
	}

	// check buffered unread data
	unreadLen := len(uc.unreadData)
	if unreadLen > 0 {
		if unreadLen <= len(b) {
			copy(b, uc.unreadData)
			uc.unreadData = make([]byte, 0)
			return unreadLen, nil
		} else {
			copy(b, uc.unreadData[0:len(b)])
			uc.unreadData = uc.unreadData[len(b):]
			return len(b), nil
		}
	}

	for {
		if err := uc.readCtx.Err(); err != nil {
			return 0, err
		}

		select {
		case <-uc.readCtx.Done():
			return 0, uc.readCtx.Err()

		case dt := <-uc.recvCh:
			if dt != nil {
				if len(dt.data) > len(b) {
					dt.data = dt.data[0:len(b)]
					n = len(b)
					uc.unreadData = dt.data[len(b):]
				} else {
					n = len(dt.data)
				}

				copy(b, dt.data)
				uc.nRecvPacket++

				uc.averageLatency = time.Duration(float64(uc.averageLatency)*(float64(uc.nRecvPacket-1)/float64(uc.nRecvPacket)) +
					float64(time.Since(dt.t))/float64(uc.nRecvPacket))

				return n, nil

			}

		}
	}
}

func (uc *UniConn) CloseWrite() error {
	uc.closeWriteCtxCancel()
	close(uc.sendCh)

	return nil
}

func (uc *UniConn) CloseRead() error {
	uc.closeReadCtxCancel()
	return nil
}

func (uc *UniConn) LocalAddr() ClientAddr {
	return ClientAddr{addr: uc.localAddr}
}

func (uc *UniConn) RemoteAddr() ClientAddr {
	return ClientAddr{addr: uc.remoteAddr}
}

func (uc *UniConn) SetDeadline(t time.Time) error {
	err := uc.SetReadDeadline(t)
	if err != nil {
		return err
	}
	err = uc.SetWriteDeadline(t)
	if err != nil {
		return err
	}
	return nil
}

func (uc *UniConn) SetReadDeadline(t time.Time) error {
	if t == zeroTime {
		uc.readCtx, uc.readCancel = context.WithCancel(uc.closeReadCtx)
	} else {
		uc.readCtx, uc.readCancel = context.WithDeadline(uc.closeReadCtx, t)
	}
	return nil
}

func (uc *UniConn) SetWriteDeadline(t time.Time) error {
	if t == zeroTime {
		uc.writeCtx, uc.writeCancel = context.WithCancel(uc.closeWriteCtx)
	} else {
		uc.writeCtx, uc.writeCancel = context.WithDeadline(uc.closeWriteCtx, t)
	}
	return nil
}

func (uc *UniConn) PrintMetrics() {
	log.Printf("%v to %v, %v packets are sent, %v packets are received, %v packets are lost, average latency is %v, loss rate is %.3f\n",
		uc.localAddr, uc.remoteAddr, uc.nSendPacket, uc.nRecvPacket, uc.nLoss, uc.averageLatency, float64(uc.nLoss)/float64(uc.nRecvPacket))
}

func (uc *UniConn) String() string {
	return fmt.Sprintf("UniConn from %v to %v", uc.localAddr, uc.remoteAddr)
}
