package core

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"encoding/json"
	"github.com/labstack/echo"
	"net"
	"time"
)

type HttpRequest struct {
	s      *Srv
	api    string
	params map[string]string
	client *http.Client
	msg    *InboxMessage
	//arduinoId string //个别arduino的命令需要转发该id
	//cardId    string //票务请求需要知道cardId与ticketId的对应关系
}

func NewHttpRequest(s *Srv) *HttpRequest {
	request := HttpRequest{}
	request.s = s
	request.api = ""
	request.client = &http.Client{
		Transport: &http.Transport{
			Dial: func(netw, addr string) (net.Conn, error) {
				conn, err := net.DialTimeout(netw, addr, time.Second*2)
				if err != nil {
					return nil, err
				}
				conn.SetDeadline(time.Now().Add(time.Second * 2))
				return conn, nil
			},
			ResponseHeaderTimeout: time.Second * 2,
		},
	}
	return &request
}

func (r *HttpRequest) SetMsg(msg *InboxMessage) {
	r.msg = msg
}

//func (r *HttpRequest) GetMsg() *InboxMessage {
//	return r.msg
//}

//func (r *HttpRequest) SetArduinoId(arduinoId string) {
//	r.arduinoId = arduinoId
//}

//func (r *HttpRequest) GetArduinoId() string {
//	return r.arduinoId
//}

func (r *HttpRequest) SetApi(api string) {
	r.api = api
}

func (r *HttpRequest) SetParams(params map[string]string) {
	r.params = params
}

func (r *HttpRequest) DoGet() {
	go func() {
		if r.api == "" {
			log.Println("http request api nil!")
			return
		}
		var httpAddr string
		u, _ := url.Parse(r.api)
		q := u.Query()
		for k, v := range r.params {
			q.Set(k, v)
		}
		u.RawQuery = q.Encode()
		httpAddr = u.String()
		log.Println("request httpAddr:", httpAddr)
		request, err := http.NewRequest(echo.GET, httpAddr, nil)
		if err != nil {
			log.Println("New request Get error:", err)
			return
		}
		request.Header.Set("Connection", "keep-alive")
		response, error := r.client.Do(request)
		if error != nil {
			log.Println("Do Get error:", error)
			hr := NewHttpResponse()
			hr.Api = r.api
			hr.Msg = r.msg
			hr.StatusCode = 408
			r.s.OnHttpRequest(hr)
			return
		}
		defer func() {
			if response != nil {
				response.Body.Close()
			}
		}()
		if response != nil {
			if response.StatusCode == http.StatusOK {
				body, _ := ioutil.ReadAll(response.Body)
				hr := NewHttpResponse()
				hr.Data = string(body)
				hr.Api = r.api
				hr.Msg = r.msg
				hr.StatusCode = http.StatusOK
				json.Unmarshal(body, &hr.JsonData)
				r.s.OnHttpRequest(hr)
			}
		}
	}()
}

func (r *HttpRequest) DoPost() {
	go func() {
		if r.api == "" {
			log.Println("http request api nil!")
			return
		}
		if r.params == nil {
			log.Println("http request params nil")
		}
		p := make(url.Values)
		for k, v := range r.params {
			p.Set(k, v)
		}
		log.Println("request httpAddr:", r.api)
		log.Println("parmas:", p)
		request, err := http.NewRequest(echo.POST, r.api, strings.NewReader(p.Encode()))
		if err != nil {
			log.Println("New request Post error:", err)
			return
		}
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		response, error := r.client.Do(request)
		if error != nil {
			log.Println("Do Post error:", error)
			hr := NewHttpResponse()
			hr.StatusCode = 408
			r.s.OnHttpRequest(hr)
			return
		}
		defer func() {
			if response != nil {
				response.Body.Close()
			}
		}()
		if response != nil {
			if response.StatusCode == 200 {
				body, _ := ioutil.ReadAll(response.Body)
				hr := NewHttpResponse()
				hr.Data = string(body)
				hr.Api = r.api
				hr.Msg = r.msg
				hr.StatusCode = http.StatusOK
				json.Unmarshal(body, &hr.JsonData)
				r.s.OnHttpRequest(hr)
			}
		}
	}()
}
