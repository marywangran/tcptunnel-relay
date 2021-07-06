package main

import (
	"io"
	"net"
	"sync"
	"fmt"
	"flag"
	"encoding/binary"
	"github.com/jamescun/tuntap"
	//"github.com/scryner/lfreequeue"
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
	wg.Add(2)

	//inq := lfreequeue.NewQueue()
	//outq := lfreequeue.NewQueue()
	go func() {
		packet := make([]byte, 2048)
		length := make([]byte, 4)

		defer wg.Done()
		for {
			n, err := io.ReadFull(conn, length[:4])
			if n == 0 || err != nil {
				tun.Close()
				break
			}
			l := binary.LittleEndian.Uint32(length)
		/*fmt.Println("read net len:%d", l)
		fmt.Println("read net len --:%d", length[0])
		fmt.Println("read net len --:%d", length[1])
		fmt.Println("read net len --:%d", length[2])
		fmt.Println("read net len --:%d", length[3])
		*/
		n, err = io.ReadFull(conn, packet[:l])
			if n == 0 || err != nil {
				break
			}
			tun.Write(packet[:l])
			//inq.Enqueue(packet[:l])
			//totun <- packet[:len]
			//tun.Write(packet[:n])
		}
	}()
	/*
	go func() {
		for {
			//pkt, _ := <-totun
			pkt, ok := inq.Dequeue()
			if ok == true {
				tun.Write(pkt.([]byte))
			}
		}
	}()
	*/
	go func() {
		packet := make([]byte, 2048)

		defer wg.Done()
		for {
			n, err := tun.Read(packet[4:])
			if n == 0 || err != nil {
				break
			}
			binary.LittleEndian.PutUint32(packet[0:4], uint32(n))
			conn.Write(packet[0:n + 4])
			//outq.Enqueue(packet[0:n + 4])
		}
	}()
	/*
	go func() {
		for {
			pkt, ok := outq.Dequeue()
			if ok == true {
				conn.Write(pkt.([]byte))
			}
		}
	}()
	*/
	wg.Wait()
}
