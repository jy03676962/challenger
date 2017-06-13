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
	conn, err := net.Dial("tcp", "172.16.10.177:4000")
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
			s = "[ID]G-1-1[TYPE]6[CARD_ID]00FF0FF000FFCF4D54B110484DBDBBB104D0[AR]1"
			//m := map[string]string{"cmd": "gameStart"}
			//b, err := json.Marshal(m)
			//if err != nil {
			//log.Println("got error:", err.Error())
			//}
			//s = string(b)
		case "2":
			s = "[ID]G-1-1[TYPE]6[CARD_ID]00FF0FF000FFCF4D54B110484DBDBBB104D0[AR]100"
			//m := map[string]string{"cmd": "nextStep"}
			//b, _ := json.Marshal(m)
			//s = string(b)
		case "3":
			//m := map[string]string{"cmd": "nextStar"}
			//b, _ := json.Marshal(m)
			//s = string(b)
			s = "[ID]G-1-1[TYPE]4[FB]4[GAME]"
			fmt.Println("print game no")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Trim(text, "\n")
			s = s + text
			fmt.Println(s)
		case "4":
			// m := map[string]string{"cmd": "addStar"}
			//b, _ := json.Marshal(m)
			//s = string(b)
			s = "[ID]G-1-1[TYPE]7[ADMIN]4[CARD_ID]00FF0FF000FFCF4D54B1104846B4FBC10480[GAME]"
			fmt.Println("print game no")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Trim(text, "\n")
			s = s + text
			fmt.Println(s)
		case "5":
			s = "[ID]G-1-1[TYPE]2[GAME]"
			fmt.Println("print game no")
			reader := bufio.NewReader(os.Stdin)
			text, _ := reader.ReadString('\n')
			text = strings.Trim(text, "\n")
			s = s + text
			fmt.Println(s)

		default:
			m := map[string]string{"cmd": text}
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
		ch1 <- "[ID]G-1-1[TYPE]0"
		<-tickChan
	}
}
