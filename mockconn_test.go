package mockconn

import (
	"fmt"
	"log"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// go test -v -run=TestBidirection
func TestBidirection(t *testing.T) {
	conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(256), Latency: 100 * time.Millisecond}
	log.Printf("Going to test bi-direction communicating, throughput is %v, latency is %v \n",
		conf.Throughput, conf.Latency)

	aliceConn, bobConn, err := NewMockConn(conf)
	require.NotNil(t, aliceConn)
	require.NotNil(t, bobConn)
	require.Nil(t, err)

	// Alice send to Bob
	nPackets := 100
	i := 0
	sendChan := make(chan []int64)
	recvChan := make(chan []int64)
	go WritePacktes(aliceConn, nPackets, sendChan)
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

	// Bob send to Alice
	nPackets = 50
	go WritePacktes(bobConn, nPackets, sendChan)
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

// go test -v -run=TestNetConnThroughput
func TestNetConnThroughput(t *testing.T) {
	tpBase := uint(256)
	for i := 1; i <= 4; i++ {
		tp := tpBase * uint(i)
		conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: tp, Latency: 20 * time.Millisecond}
		log.Printf("Going to test throughput at %v packets/s, latency %v\n", conf.Throughput, conf.Latency)

		aliceConn, bobConn, err := NewMockConn(conf)
		require.NotNil(t, aliceConn)
		require.NotNil(t, bobConn)
		require.Nil(t, err)

		nPackets := 256
		sendChan := make(chan []int64)
		recvChan := make(chan []int64)
		go WritePacktes(aliceConn, nPackets, sendChan)
		go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

		<-sendChan
		<-recvChan
		fmt.Println()
	}
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
	go WritePacktes(aliceConn, nPackets, sendChan)
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
	go WritePacktes(aliceConn, nPackets, sendChan)
	go ReadPackets(bobConn, nPackets, conf.Latency, recvChan)

	<-sendChan
	aliceConn.Close()
	bobConn.Close()
	<-recvChan
}
