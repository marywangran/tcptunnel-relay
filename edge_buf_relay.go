package main

// 为了缩短处理路径，用一个lockless队列来隔离生产者和消费者
import (
	"io"
	"net"
	"sync"
	"fmt"
	"flag"
	"encoding/binary"
	"github.com/jamescun/tuntap"
	"github.com/scryner/lfreequeue"
)

type Packet []byte

func main() {
	var wg sync.WaitGroup
	var host string
	var port int
	flag.StringVar(&host, "h", "1.1.1.1", "host")
	flag.IntVar(&port, "p", 1234, "port")
	flag.Parse()
	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr {
		IP:   net.ParseIP(host),
		Port: port,
		Zone: "",
	})
	if err != nil {
		return
	}
	defer conn.Close()

	tun, err := tuntap.Tun("edge")
	if err != nil {
		fmt.Println("error: tun:", err)
		return
	}
	defer tun.Close()
	//totun := make(chan Packet, 20000)
	wg.Add(4)

	inq := lfreequeue.NewQueue()
	outq := lfreequeue.NewQueue()
	go func() {
		packet := make([]byte, 2048)
		length := make([]byte, 4)

		defer wg.Done()
		for {
			n, err := io.ReadFull(conn, length)
			if n == 0 || err != nil {
				tun.Close()
				break
			}
			len := binary.LittleEndian.Uint32(length)
			n, err = io.ReadFull(conn, packet[:len])
			if n == 0 || err != nil {
				break
			}
			inq.Enqueue(packet) // 用别人实现的一个lockless队列
			//totun <- packet[:len] // 用channel性能非常差！
			//tun.Write(packet[:n])
		}
	}()
	go func() {
    defer wg.Done()
		for {
			//pkt, _ := <-totun
			pkt, ok := inq.Dequeue()
			if ok == true {
        _, err := tun.Write(pkt.([]byte))
        if err != nil {
          break
        }
			}
		}
	}()
	go func() {
		packet := make([]byte, 2048)

		defer wg.Done()
		for {
			len, err := tun.Read(packet)
			if len == 0 || err != nil {
				break
			}
			outq.Enqueue(packet) // 缩短处理路径，放入queue后直接返回
			//binary.LittleEndian.PutUint32(length, uint32(len))
			//conn.Write(length)
			//conn.Write(packet[:len])
		}
	}()
	go func() {
    defer wg.Done()
		length := make([]byte, 4)
		for {
			pkt, ok := outq.Dequeue()
			if ok == true {
				binary.LittleEndian.PutUint32(length, uint32(len(pkt.([]byte))))
        conn.Write(length)
        _, err := conn.Write(pkt.([]byte))
        if err != nil {
          break
        }
			}
		}
	}()
	wg.Wait()
}
