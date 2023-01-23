package mockconn

import (
	"encoding/binary"
	"log"
	"net"
	"time"
)

func WritePacktes(writer net.Conn, nPackets int, sendCh chan []int64) {
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

	sendCh <- sendSeq // return the sequences written
}

func ReadPackets(reader net.Conn, nPackets int, latency time.Duration, recvCh chan []int64) {
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

	if uc, ok := reader.(*UniConn); ok {
		uc.PrintMetrics()
	} else if nc, ok := reader.(*NetConn); ok {
		nc.PrintMetrics()
	}

	recvCh <- recvSeq // return the sequences read
}
