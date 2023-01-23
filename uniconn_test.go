package mockconn

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// go test -v -run=TestUniConnThroughput
func TestUniConnThroughput(t *testing.T) {
	tpBase := uint(256)
	for i := 1; i <= 4; i++ {
		tp := tpBase * (uint)(i)
		conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: tp, Latency: 20 * time.Millisecond}
		uc, err := NewUniConn(conf)
		require.Nil(t, err)
		require.NotNil(t, uc)

		fmt.Println("target tp is", tp)
		sendChan := make(chan []int64)
		recvChan := make(chan []int64)
		nPacket := 100
		go WritePacktes(uc, nPacket, sendChan)
		go ReadPackets(uc, nPacket, conf.Latency, recvChan)

		<-sendChan
		<-recvChan
	}
}

// go test -v -run=TestSetReadDeadline
func TestSetReadDeadline(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(16), Latency: 100 * time.Millisecond}
	uc, err := NewUniConn(conf)
	require.Nil(t, err)
	require.NotNil(t, uc)

	uc.SetReadDeadline(time.Now().Add(time.Second))
	b := make([]byte, 1024)
	n, err := uc.Read(b)
	require.NotNil(t, err)
	require.Equal(t, 0, n)

	n, err = uc.Write(b)
	require.Nil(t, err)
	require.Equal(t, 1024, n)
	n, err = uc.Read(b)
	require.Nil(t, err)
	require.Equal(t, 1024, n)
}

// go test -v -run=TestClose
func TestClose(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(16), Latency: 100 * time.Millisecond}
	uc, err := NewUniConn(conf)
	require.Nil(t, err)
	require.NotNil(t, uc)

	b := []byte("hello")
	n, err := uc.Write(b)
	require.Nil(t, err)
	require.Equal(t, len(b), n)

	b2 := make([]byte, 1024)
	n, err = uc.Read(b2)

	require.Nil(t, err)
	require.Equal(t, len(b), n)

	uc.Write(b)
	uc.CloseRead()
	n, err = uc.Read(b2)
	require.NotNil(t, err)
	require.Equal(t, 0, n)
	t.Log("After close read, read err ", err)

	uc.CloseWrite()
	n, err = uc.Write(b)
	require.NotNil(t, err)
	require.Equal(t, 0, n)
	t.Log("After close write, write err ", err)
}

// go test -v -run=TestRateLimiter
func TestRateLimiter(t *testing.T) {
	lim := 2000
	r := rate.NewLimiter(rate.Limit(lim), 1)
	count := 10000

	start := time.Now()
	for i := 0; i < count; i++ {
		r.Wait(context.Background())
	}
	d := time.Since(start)

	fmt.Printf("Count %v took %v, average is %.1f, expected is %v\n",
		count, d, float64(count)/d.Seconds(), lim)
}

// go test -v -run=TestCloseRead
func TestCloseRead(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(16), Latency: 50 * time.Millisecond, WriteTimeout: 3 * time.Second}
	uc, err := NewUniConn(conf)
	require.Nil(t, err)
	require.NotNil(t, uc)

	writeCh := make(chan struct{})
	go func() {
		i := 0
		for i = 0; i < 1000; i++ {
			b := make([]byte, 1024)
			n, err := uc.Write(b)
			if err != nil {
				t.Log("write err", err)
				break
			} else {
				require.Equal(t, n, len(b))
			}
		}
		close(writeCh)
	}()

	go func() {
		i := 0
		for i = 0; i < 1000; i++ {
			b := make([]byte, 1024)
			n, err := uc.Read(b)
			if err != nil {
				t.Log("read err", err)
				break
			} else {
				require.Equal(t, n, len(b))
			}

			if i == 100 {
				err := uc.CloseRead()
				if err != nil {
					t.Log("CloseRead err", err)
				}
			}
		}
		require.Equal(t, 101, i)
	}()

	<-writeCh
}

// go test -v -run=TestTimeout
func TestTimeout(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(16), Latency: 20 * time.Millisecond,
		WriteTimeout: 2 * time.Second, ReadTimeout: 3 * time.Second}
	uc, err := NewUniConn(conf)
	require.Nil(t, err)
	require.NotNil(t, uc)

	for i := 0; i < 1000; i++ {
		b := []byte("hello")
		n, err := uc.Write(b)
		if err != nil {
			require.Equal(t, 0, n)
			require.Equal(t, context.DeadlineExceeded, err)
			break
		}
	}

	for i := 0; i < 1000; i++ {
		b := make([]byte, 1024)
		n, err := uc.Read(b)
		if err != nil {
			require.Equal(t, 0, n)
			require.Equal(t, context.DeadlineExceeded, err)
			break
		}
	}
}
