package main

import (
	"io"
	"net"
	"sync"
	"fmt"
	"flag"
	"encoding/binary"
	"github.com/jamescun/tuntap"
)

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
	wg.Add(2)

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
			tun.Write(packet[:n])
		}
	}()
	go func() {
		packet := make([]byte, 2048)
		length := make([]byte, 4)

		defer wg.Done()
		for {
			len, err := tun.Read(packet)
			if len == 0 || err != nil {
				break
			}
			binary.LittleEndian.PutUint32(length, uint32(len))
			conn.Write(length)
			conn.Write(packet[:len])
		}
	}()
	wg.Wait()
}
