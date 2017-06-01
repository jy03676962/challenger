package core

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/labstack/echo"
)

type HttpRequest struct {
	s         *Srv
	api       string
	params    map[string]string
	client    *http.Client
	arduinoId string //个别arduino的命令需要转发该id
	cardId    string //票务请求需要知道cardId与ticketId的对应关系
	opration  string
}

func NewHttpRequest(s *Srv) *HttpRequest {
	request := HttpRequest{}
	request.s = s
	request.api = ""
	request.client = &http.Client{}
	return &request
}

func (r *HttpRequest) SetCardId(cardId string) {
	r.cardId = cardId
}

//func (r *HttpRequest) GetCardId() string {
//	return r.cardId
//}

func (r *HttpRequest) SetArduinoId(arduinoId string) {
	r.arduinoId = arduinoId
}

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
		request, _ := http.NewRequest(echo.GET, httpAddr, nil)
		request.Header.Set("Connection", "keep-alive")
		response, _ := r.client.Do(request)
		defer func() {
			if response != nil {
				response.Body.Close()
			}
		}()
		if response != nil {
			if response.StatusCode == 200 {
				body, _ := ioutil.ReadAll(response.Body)
				hr := HttpResponse{}
				hr.Data = string(body)
				hr.Api = r.api
				if r.arduinoId != "" {
					hr.ArduinoId = r.arduinoId
				}
				if r.cardId != "" {
					hr.CardId = r.cardId
				}
				r.s.OnHttpRequest(&hr)
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
		log.Println(p)
		request, _ := http.NewRequest(echo.POST, r.api, strings.NewReader(p.Encode()))
		request.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
		response, _ := r.client.Do(request)
		defer func() {
			if response != nil {
				response.Body.Close()
			}
		}()
		if response != nil {
			if response.StatusCode == 200 {
				body, _ := ioutil.ReadAll(response.Body)
				hr := HttpResponse{}
				hr.Data = string(body)
				hr.Api = r.api
				if r.arduinoId != "" {
					hr.ArduinoId = r.arduinoId
				}
				if r.cardId != "" {
					hr.CardId = r.cardId
				}
				r.s.OnHttpRequest(&hr)
			}
		}
	}()
}
