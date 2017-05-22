package core

import "log"

var _ = log.Printf
type LoginInfo struct {
	PlayerNum int
	PlayerCardInfo map[int]string
}

//占卜
type Adivainacion struct {
	GameId     int
	Card_ID    string
	Time_start string
	Time_end   string
}

func (game * Adivainacion)Reset()  {
	game.GameId = 0
	game.Card_ID = ""
	game.Time_start = ""
	game.Time_end = ""
}

//六连
type Bang struct {
	GameId      int
	Card_ID     string
	Time_start  string
	Time_end    string
	Point_round map[int]int //设定为3局map[局数]分数
}

func (game * Bang)Reset()  {
	game.GameId = 0
	game.Card_ID = ""
	game.Time_start = ""
	game.Time_end = ""
	for k,_ := range game.Point_round {
		delete(game.Point_round,k)
	}
}

//走格子
type Follow struct {
	GameId     int
	Card_ID1   string
	Card_ID2   string
	Time_start string
	Time_end   string
	Last_round int
}

func (game * Follow)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Last_round = 0
}

//新人走廊
type Greeting struct {
	GameId     int
	Card_ID1   string
	Card_ID2   string
	Time_start string
	Time_end   string
}

func (game * Greeting)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
}

//午时已到
type Highnoon struct {
	GameId          int
	Card_ID1        string
	Card_ID2        string
	Time_start      string
	Time_end        string
	Result_round_1p map[int]float64 //共7局 map[7]0.617 float代表开枪时间
	Result_round_2p map[int]float64 //共7局 map[7]0.617 float代表开枪时间
}

func (game * Highnoon)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	for k,_ := range game.Result_round_1p {
		delete(game.Result_round_1p,k)
	}
	for k,_ := range game.Result_round_2p {
		delete(game.Result_round_2p,k)
	}
}

//寻宝
type Hunter struct {
	GameId           int
	Card_ID1         string
	Card_ID2         string
	Time_start       string
	Time_end         string
	Time_firstButton string
	Box_ID           int
}

func (game * Hunter)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Time_firstButton = ""
	game.Box_ID = 0
}

//寻宝所分配的场地保箱
type HunterBox struct {
	Box_ID        int
	Time_build    string
	Time_validity string
	Card_ID1      string
	Card_ID2      string
	Box_status    int //0代表未开启，1代表开启
	IsAssigned    bool
}

//射箭
type Marksman struct {
	GameId      int
	Card_ID1    string
	Card_ID2    string
	Time_start  string
	Time_end    string
	Point_left  int
	Point_right int
}

func (game * Marksman)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Point_left = 0
	game.Point_right = 0
}

//挖矿
type Miner struct {
	GameId     int
	Card_ID1   string
	Card_ID2   string
	Time_start string
	Time_end   string
}

func (game * Miner)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
}

//默契牢笼
type Privity struct {
	GameId       int
	Card_ID1     string
	Card_ID2     string
	Time_start   string
	Time_end     string
	Num_question int
	Num_right    int
}

func (game * Privity)Reset()  {
	game.GameId = 0
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Num_question = 0
	game.Num_right = 0
}

//献祭房间
type Russian struct {
	GameId     int
	Card_ID    string
	Time_start string
	Time_end   string
}

func (game * Russian)Reset()  {
	game.GameId = 0
	game.Card_ID = ""
	game.Time_start = ""
	game.Time_end = ""
}
