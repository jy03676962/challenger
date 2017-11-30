package main

import (
	"challenger/server/core"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/labstack/echo"
	st "github.com/labstack/echo/engine/standard"
	mw "github.com/labstack/echo/middleware"
)
//Timerift 3.0
const (
	host        = "192.168.1.5"
	httpAddr    = host + ":3000"
	tcpAddr     = host + ":4000"
	adminAddr   = host + ":5000"
	dbPath      = "./challenger.db"
	isSimulator = false
	testRank    = true
)

func redirectStderr(f *os.File) {
	err := syscall.Dup2(int(f.Fd()), int(os.Stderr.Fd()))
	if err != nil {
		log.Fatalf("Failed to redirect stderr to file: %v", err)
	}
}

func loadRankTestData() map[string]interface{} {
	ret := make(map[string]interface{})
	b, e := ioutil.ReadFile("./ranktest.json")
	if e != nil {
		log.Println("load rank test data error:%v\n", e)
		os.Exit(1)
	}
	e = json.Unmarshal(b, &ret)
	if e != nil {
		log.Println("parse rank test data error:%v\n", e)
		os.Exit(1)
	}
	return ret
}

func main() {
	// setup log system
	log.Println("start server")
	logfileName := "log/" + time.Now().Local().Format("2006-01-02-15-04-05") + ".log"
	os.Mkdir("log", 0777)
	f, err := os.OpenFile(logfileName, os.O_WRONLY|os.O_CREATE, 0640)
	if err != nil {
		fmt.Println("error open log file", err)
		os.Exit(1)
	}
	if runtime.GOOS != "windows" {
		pf, err := os.OpenFile("panic.log", os.O_WRONLY|os.O_CREATE, 0640)
		if err != nil {
			fmt.Println("error open panic file", err)
			os.Exit(1)
		}
		redirectStderr(pf)
	}
	log.SetOutput(io.MultiWriter(f, os.Stdout))
	log.Println("setup log system done")

	defer func() {
		if err := recover(); err != nil { //catch
			log.Printf("Exception: %v\n", err)
			os.Exit(1)
		}
	}()

	var rankTestData map[string]interface{}

	if testRank {
		rankTestData = loadRankTestData()
	}

	core.GetOptions()

	log.Println("reading cfg done")

	srv := core.NewSrv(isSimulator)
	go srv.Run(tcpAddr, adminAddr, dbPath)

	// setup echo
	ec := echo.New()
	ec.Static("/", "public")
	ec.Static("/api/asset/", "api_public")
	ec.Use(mw.Logger())
	ec.Get("/api/allhistory", func(c echo.Context) error {
		if rankTestData == nil {
			return c.JSON(http.StatusOK, nil)
		}
		data := make(map[string]interface{})
		data["mode0"] = rankTestData["gold"]
		data["mode1"] = rankTestData["survival"]
		data["code"] = "0"
		data["error"] = ""
		return c.JSON(http.StatusOK, data)
	})
	ec.Post("/api/mode1history", func(c echo.Context) error {
		if rankTestData == nil {
			return c.JSON(http.StatusOK, nil)
		}
		month, _ := strconv.Atoi(c.FormValue("month"))
		year, _ := strconv.Atoi(c.FormValue("year"))
		data := rankTestData["survival"].(map[string]interface{})
		data["season"] = fmt.Sprintf("S%d02%d", year, month)
		data["code"] = "0"
		data["error"] = ""
		return c.JSON(http.StatusOK, data)
	})
	log.Println("listen http:", httpAddr)
	ec.Run(st.New(httpAddr))
}
