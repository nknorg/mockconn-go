# mockconn

Mock a network connection which has two endpoints to communicate.
The connection implement of go standard library **net.Conn** interface

This network connection has configurable parameters as :

```
type ConnConfig struct {
    Addr1      string
    Addr2      string
    Throughput uint
    BufferSize uint
    Latency    time.Duration
    Loss       float32 // 0.01 = 1%
}
```

* Addr1: Address or any name to identify one endpoint, such as "Alice" or "127.0.0.1"
* Addr2: Address or any name to identify the other endpoint, such as "Bob" or an IP address
* Throughput: The Throughput (packet/second) you set for this connection. Each packet is default to 1024 bytes.
* BufferSize: The buffer size used in the network. It is suggest equal or greater than throughput.
* Latency: The duration which a packet travels in the connection. 
* Loss: The rate of loss in the connection.

You can mock out a connection as:

```
conf := &ConnConfig{Addr1: "Alice", Addr2: "Bob", Throughput: uint(256), Latency: 100 *     time.Millisecond}
aliceConn, bobConn, err := NewMockConn(conf)
```

After mocking out the connection. You can begin send and receive data as:

* Writer

```
b := make([]byte, 1024)
// Todo: put meaningful data into b
n, err := aliceConn.Write(b)
if err != nil {
    return err
}
```

* Reader

```
b := make([]byte, 1024)
n, err := bobConn.Read(b)
if err != nil {
    return ree
}
```

After fininshing data transmitting and receiving all data, you may close this connection at both endpoints:

```
aliceConn.Close()
bobConn.Close()
```
