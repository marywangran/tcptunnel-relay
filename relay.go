package main

import (
	"io"
	"sync"
	"flag"
	"log"
	"net"
)

func main() {
	var host string
	var port int
	var mode int
	var base int
	flag.StringVar(&host, "h", "10.198.54.78", "host")
	flag.IntVar(&port, "p", 1234, "port 1234")
	flag.IntVar(&mode, "m", 0, "server|server-client")
	flag.Parse()

	base = port
	if mode == 0 {
		listener, err := net.ListenTCP("tcp", &net.TCPAddr{
			IP:   net.ParseIP(host),
			Port: port,
			Zone: "",
		})
		if err != nil {
			log.Panic(err)
		}
		log.Printf("listen:%d\n", port)
		for {
			front, err := listener.Accept()
			if err != nil {
				log.Fatal(err)
			}
			port = port + 1
			log.Printf("front:%d  back:%d\n", base, port)
			go handle_front(front, port)
		}
	} else {
		front, err := net.DialTCP("tcp", nil, &net.TCPAddr {
			IP:   net.ParseIP(host),
			Port: port,
			Zone: "",
		})
		if err != nil {
			return
		}
		handle_front(front, port)
	}
}

func handle_front(front net.Conn, port int) {
	defer front.Close()

	blistener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("0.0.0.0"),
		Port: port,
		Zone: "",
	})
	if err != nil {
		log.Panic(err)
	}
	back, err := blistener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	defer back.Close()

	var wg sync.WaitGroup

	wg.Add(2) 
	// 全双工裸转发，保持核心的简单，复杂丢给端到端处理
	go func() {
		defer wg.Done()
		io.Copy(front, back)
	}()
	go func() {
		defer wg.Done()
		io.Copy(back, front)
	}()
	wg.Wait()
}
