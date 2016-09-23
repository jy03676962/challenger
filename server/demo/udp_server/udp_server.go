package main

import (
	"bufio"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	HOST_PORT = "5000"
	// GUEST_PORT = "8989"
)

type DeviceData struct {
	dict map[string]*net.UDPAddr
	conn *net.UDPConn
	lock *sync.RWMutex
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Exception: %v\n", err)
		}
	}()
	cmdChan := make(chan string, 1)
	go start_udp_server(cmdChan)
	// go start_mock_udp_client()
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		cmdChan <- strings.Trim(text, "\n")
	}

}

// func start_mock_udp_client() {
// 	serverAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+HOST_PORT)
// 	if err != nil {
// 		fmt.Println("Resolve Server Adress Error", err.Error())
// 		os.Exit(1)
// 	}
// 	localAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:"+GUEST_PORT)
// 	if err != nil {
// 		fmt.Println("Resolve Client Local Adress Error", err.Error())
// 		os.Exit(1)
// 	}
// 	conn, err := net.DialUDP("udp", localAddress, serverAddress)
// 	if err != nil {
// 		fmt.Println("Dial UDP Error", err.Error())
// 		os.Exit(1)
// 	}
// 	defer conn.Close()
// 	go func() {
// 		buf := make([]byte, 1024)
// 		for {
// 			n, addr, err := conn.ReadFromUDP(buf)
// 			message := string(buf[0:n])
// 			fmt.Println("client received", message, "from", addr)
// 			if err != nil {
// 				fmt.Println("read server error", err)
// 			}
// 		}
// 	}()
// 	for {
// 		_, err := conn.Write([]byte("HBT001"))
// 		if err != nil {
// 			fmt.Println("sent to server error", err.Error())
// 		}
// 		time.Sleep(time.Second * 3)
// 	}
// }

func start_udp_server(cmdChan chan string) {
	d := DeviceData{}
	d.dict = make(map[string]*net.UDPAddr)
	d.lock = new(sync.RWMutex)
	go start_listen_cmd(cmdChan, &d)
	start_listen_UDP(&d)
}

func start_listen_cmd(cmdChan chan string, d *DeviceData) {
	for {
		select {
		case message := <-cmdChan:
			log.Printf("Got cmd:%v, len:%v\n", message, len(message))
			if len(message) > 6 {
				deviceId := message[3:6]
				var conn *net.UDPConn
				d.lock.RLock()
				addr, prs := d.dict[deviceId]
				if prs {
					conn = d.conn
				}
				d.lock.RUnlock()
				if conn != nil && addr != nil {
					_, err := conn.WriteToUDP([]byte(message), addr)
					if err != nil {
						log.Println("write to client error", err)
					} else {
						log.Println(time.Now().String(), "send cmd to client:", message)
					}
				}
			}
		}
	}
}

func start_listen_UDP(d *DeviceData) {
	UDPAddress, err := net.ResolveUDPAddr("udp", "192.168.1.5:"+HOST_PORT)
	if err != nil {
		log.Println("Resolve Server Local Adress Error", err.Error())
		os.Exit(1)
	}
	UDPConn, err := net.ListenUDP("udp", UDPAddress)
	if err != nil {
		log.Println("Listen UDP Error", err.Error())
		os.Exit(1)
	}
	d.lock.Lock()
	d.conn = UDPConn
	d.lock.Unlock()
	defer UDPConn.Close()
	buf := make([]byte, 1024)
	for {
		n, addr, err := UDPConn.ReadFromUDP(buf)
		message := string(buf[0:n])
		log.Println(time.Now().String(), "server received", message, "from", addr)
		if err != nil {
			log.Println("Read Error", err)
		}
		if len(message) >= 6 {
			deviceId := message[3:6]
			d.lock.Lock()
			d.dict[deviceId] = addr
			d.lock.Unlock()
		}
	}
}
