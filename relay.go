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
	var base int
	flag.StringVar(&host, "h", "10.198.54.78", "host")
	flag.IntVar(&port, "p", 1234, "port 1234")
	flag.Parse()

	base = port
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
		go handle_front(front, host, port)
	}
}

func handle_front(front net.Conn, addr string, port int) {
	defer front.Close()

	blistener, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP(addr),
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
