package mockconn

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/time/rate"
)

// go test -v -run=TestSetReadDeadline
func TestSetReadDeadline(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(16), Latency: 100 * time.Millisecond}
	uc, err := NewUniConn(conf)
	require.Nil(t, err)
	require.NotNil(t, uc)

	uc.SetDeadline(time.Now().Add(time.Second))
	b := make([]byte, 1500)
	n, err := uc.Read(b)
	require.Equal(t, 0, n)
	t.Log("Read with deadline, err: ", err)
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
