package mockconn

import (
	"encoding/binary"
	"log"
	"math"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func WritePacktes(writer net.Conn, nPackets int, latency time.Duration, ch chan []int64) {
	var sendSeq []int64

	seq := int64(1)
	count := int64(0)
	t1 := time.Now()
	for i := 0; i < nPackets; i++ {
		b := make([]byte, 1024)
		binary.PutVarint(b, seq)
		_, err := writer.Write(b)
		if err != nil {
			break
		}
		sendSeq = append(sendSeq, seq)
		seq++
		count++
	}
	dur := time.Since(t1)
	throughput := float64(count) / dur.Seconds()

	log.Printf("%v send %v packets in %v ms, throughput is: %.1f packets/s \n",
		writer.LocalAddr(), count, dur.Milliseconds(), throughput)

	ch <- sendSeq

	return
}

func ReadPackets(reader net.Conn, nPackets int, latency time.Duration, ch chan []int64) {
	var recvSeq []int64

	b := make([]byte, 1024)
	count := int64(0)
	t1 := time.Now()
	for i := 0; i < nPackets; i++ {
		_, err := reader.Read(b)
		if err != nil {
			break
		}
		seq, _ := binary.Varint(b[:8])
		recvSeq = append(recvSeq, seq)
		count++
	}
	dur := time.Since(t1)
	dur2 := dur - latency
	throughput := float64(count) / dur2.Seconds()

	log.Printf("%v read %v packets in %v ms, deduct %v ms latency, throughput is: %.1f packets/s \n",
		reader.LocalAddr(), count, dur.Milliseconds(), latency.Milliseconds(), throughput)
	nc, ok := reader.(*NetConn)
	if ok {
		nc.PrintMetrics()
	}

	ch <- recvSeq

	return
}

// go test -v -run=TestBidirection
func TestBidirection(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(256), Latency: 100 * time.Millisecond}
	log.Printf("Going to test bi-direction communicating, throughput is %v, latency is %v \n",
		conf.Throughput, conf.Latency)

	aliceConn, bobConn, err := NewMockConn(conf)
	require.NotNil(t, aliceConn)
	require.NotNil(t, bobConn)
	require.Nil(t, err)

	// left write to right
	nPackets := 100
	i := 0
	sendChan := make(chan []int64)
	recvChan := make(chan []int64)
	go WritePacktes(aliceConn, nPackets, conf.Latency, sendChan)
	go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

	sendSeq := <-sendChan
	recvSeq := <-recvChan

	minLen := int(math.Min(float64(len(sendSeq)), float64(len(recvSeq))))
	for i = 0; i < minLen; i++ {
		if sendSeq[i] != recvSeq[i] {
			log.Printf("%v sendSeq[%v] %v != %v recvSeq[%v] %v \n",
				aliceConn.LocalAddr(), i, sendSeq[i], bobConn.LocalAddr(), i, recvSeq[i])
		}
	}
	if i == nPackets {
		log.Printf("%v write to %v %v packets, %v receive %v packets in the same sequence\n",
			aliceConn.LocalAddr(), bobConn.LocalAddr(), nPackets, bobConn.LocalAddr(), nPackets)
	}

	// right write to left
	nPackets = 50
	go WritePacktes(bobConn, nPackets, conf.Latency, sendChan)
	go ReadPackets(aliceConn, nPackets, conf.Latency, recvChan)

	sendSeq = <-sendChan
	recvSeq = <-recvChan

	minLen = int(math.Min(float64(len(sendSeq)), float64(len(recvSeq))))
	for i = 0; i < minLen; i++ {
		if sendSeq[i] != recvSeq[i] {
			log.Printf("%v sendSeq[%v] %v != %v recvSeq[%v] %v \n",
				bobConn.LocalAddr(), i, sendSeq[i], aliceConn.LocalAddr(), i, recvSeq[i])
		}
	}
	if i == nPackets {
		log.Printf("%v write to %v %v packets, %v receive %v packets in the same sequence\n",
			bobConn.LocalAddr(), aliceConn.LocalAddr(), nPackets, aliceConn.LocalAddr(), nPackets)
	}
}

// go test -v -run=TestLowThroughput
func TestLowThroughput(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(16), Latency: 20 * time.Millisecond}
	log.Printf("Going to test low throughput at %v packets/s, latency %v\n", conf.Throughput, conf.Latency)

	aliceConn, bobConn, err := NewMockConn(conf)
	require.NotNil(t, aliceConn)
	require.NotNil(t, bobConn)
	require.Nil(t, err)

	nPackets := 256
	sendChan := make(chan []int64)
	recvChan := make(chan []int64)
	go WritePacktes(aliceConn, nPackets, conf.Latency, sendChan)
	go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

	<-sendChan
	<-recvChan
}

// go test -v -run=TestHighThroughput
func TestHighThroughput(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(2048), Latency: 20 * time.Millisecond}
	log.Printf("Going to test high throughput at %v packets/s, latency %v\n", conf.Throughput, conf.Latency)

	aliceConn, bobConn, err := NewMockConn(conf)
	require.NotNil(t, aliceConn)
	require.NotNil(t, bobConn)
	require.Nil(t, err)

	nPackets := 2048
	sendChan := make(chan []int64)
	recvChan := make(chan []int64)
	go WritePacktes(aliceConn, nPackets, conf.Latency, sendChan)
	go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

	<-sendChan
	<-recvChan
}

// go test -v -run=TestHighLatency
func TestHighLatency(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(128), Latency: 500 * time.Millisecond}
	log.Printf("Going to test throughput at %v packets/s, high latency %v\n", conf.Throughput, conf.Latency)

	aliceConn, bobConn, err := NewMockConn(conf)
	require.NotNil(t, aliceConn)
	require.NotNil(t, bobConn)
	require.Nil(t, err)

	nPackets := 256
	sendChan := make(chan []int64)
	recvChan := make(chan []int64)
	go WritePacktes(aliceConn, nPackets, conf.Latency, sendChan)
	go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

	<-sendChan
	<-recvChan
}

// go test -v -run=TestLoss
func TestLoss(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(128), Latency: 20 * time.Millisecond, Loss: 0.01}
	log.Printf("Going to test throughput at %v packets/s, latency %v, loss :%v \n", conf.Throughput, conf.Latency, conf.Loss)

	aliceConn, bobConn, err := NewMockConn(conf)
	require.NotNil(t, aliceConn)
	require.NotNil(t, bobConn)
	require.Nil(t, err)

	nPackets := 256
	sendChan := make(chan []int64)
	recvChan := make(chan []int64)
	go WritePacktes(aliceConn, nPackets, conf.Latency, sendChan)
	go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

	<-sendChan
	aliceConn.Close()
	bobConn.Close()
	<-recvChan
}
