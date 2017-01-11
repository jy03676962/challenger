package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "192.168.1.5:5000")
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	ch := make(chan string, 1)
	//go read(conn, ch)
	go write(conn, ch)
	ch1 := make(chan string, 1)
	go writeHeart(conn, ch1)
	go writech1(ch1)
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		text = strings.Trim(text, "\n")
		var s string
		switch text {
		case "0":
			//m := map[string]string{"cmd": "upload_score", "score": "A"}
			m := map[string]string{"cmd": "init"}
			b, err := json.Marshal(m)
			if err != nil {
				log.Println("got error:", err.Error())
			}
			s = string(b)
		case "1":
			//s = "[UR]100000000111111"
			m := map[string]string{"cmd": "gameStart"}
			b, err := json.Marshal(m)
			if err != nil {
				log.Println("got error:", err.Error())
			}
			s = string(b)
		case "2":
			m := map[string]string{"cmd": "nextStep"}
			b, _ := json.Marshal(m)
			s = string(b)
		case "3":
			m := map[string]string{"cmd": "confirm_init_score"}
			b, _ := json.Marshal(m)
			s = string(b)
		}
		ch <- s
	}

}

func read(conn net.Conn, ch chan string) {
	r := bufio.NewReader(conn)
	for {
		b, err := r.ReadByte()
		if err != nil {
			log.Println("err:", err.Error())
			os.Exit(1)
		}
		if b != 60 {
			log.Println("hardware message must start with <")
			os.Exit(1)
		}
		msg := make([]byte, 0)
		for {
			b, err := r.ReadByte()
			if err != nil {
				log.Println("err:", err.Error())
				os.Exit(1)
			}
			if b == 62 {
				break
			}
			msg = append(msg, b)
		}
		if len(msg) == 0 {
			log.Println("got empty hardware message")
			os.Exit(1)
		}
		msgStr := string(msg)
		log.Println("got message: ", msgStr)
	}
}

func write(conn net.Conn, ch chan string) {
	for {
		s := <-ch
		fmt.Fprintf(conn, "<"+s+">")
	}
}

func writeHeart(conn net.Conn, ch chan string) {
	for {
		s := <-ch
		fmt.Fprintf(conn, "<"+s+">")
	}
}

func writech1(ch1 chan string) {
	dt := 500 * time.Millisecond
	tickChan := time.Tick(dt)
	for {
		ch1 <- "[ID]Admin"
		<-tickChan
	}
}
