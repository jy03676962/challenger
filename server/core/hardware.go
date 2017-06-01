package core

import "log"

var _ = log.Printf

const (
	ID_Russian = iota //轮盘赌
	ID_Adivainacion //占卜
	ID_Bang //6连
	ID_Follow //走格子
	ID_Greeting //新人走廊
	ID_Highnoon //午时已到
	ID_Hunter //寻宝
	ID_Marksman //射箭
	ID_Miner //挖矿
	ID_Privity //默契牢笼
)

type LoginInfo struct {
	PlayerNum      int
	PlayerCardInfo map[string]string //1p:cardId
	CardTicketInfo map[string]string //cardId:ticketId
}

//占卜
type Adivainacion struct {
	GameId     int
	Card_ID    string
	Time_start string
	Time_end   string
	LoginInfo  *LoginInfo
}

func NewAdivainacion() *Adivainacion {
	game := Adivainacion{}
	game.GameId = ID_Adivainacion
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Adivainacion) Reset() {
	game.Card_ID = ""
	game.Time_start = ""
	game.Time_end = ""
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
}

//六连
type Bang struct {
	GameId      int
	Card_ID     string
	Time_start  string
	Time_end    string
	Point_round map[int]int //设定为3局map[局数]分数
	LoginInfo   *LoginInfo
}

func NewBang() *Bang {
	game := Bang{}
	game.GameId = ID_Bang
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Bang) Reset() {
	game.Card_ID = ""
	game.Time_start = ""
	game.Time_end = ""
	for k, _ := range game.Point_round {
		delete(game.Point_round, k)
	}
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
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
	LoginInfo  *LoginInfo
}

func NewFollow() *Follow {
	game := Follow{}
	game.GameId = ID_Follow
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Follow) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Last_round = 0
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
}

//新人走廊
type Greeting struct {
	GameId     int
	Card_ID1   string
	Card_ID2   string
	Time_start string
	Time_end   string
	LoginInfo  *LoginInfo
}

func NewGreeting() *Greeting {
	game := Greeting{}
	game.GameId = ID_Greeting
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Greeting) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
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
	LoginInfo       *LoginInfo
}

func NewHighnoon() *Highnoon {
	game := Highnoon{}
	game.GameId = ID_Highnoon
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Highnoon) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	for k, _ := range game.Result_round_1p {
		delete(game.Result_round_1p, k)
	}
	for k, _ := range game.Result_round_2p {
		delete(game.Result_round_2p, k)
	}
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
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
	LoginInfo        *LoginInfo
}

func NewHunter() *Hunter {
	game := Hunter{}
	game.GameId = ID_Hunter
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Hunter) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Time_firstButton = ""
	game.Box_ID = 0
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
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
	LoginInfo   *LoginInfo
}

func NewMarksman() *Marksman {
	game := Marksman{}
	game.GameId = ID_Marksman
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Marksman) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Point_left = 0
	game.Point_right = 0
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
}

//挖矿
type Miner struct {
	GameId     int
	Card_ID1   string
	Card_ID2   string
	Time_start string
	Time_end   string
	LoginInfo  *LoginInfo
}

func NewMiner() *Miner {
	game := Miner{}
	game.GameId = ID_Miner
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Miner) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
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
	LoginInfo    *LoginInfo
}

func NewPrivity() *Privity {
	game := Privity{}
	game.GameId = ID_Privity
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Privity) Reset() {
	game.Card_ID1 = ""
	game.Card_ID2 = ""
	game.Time_start = ""
	game.Time_end = ""
	game.Num_question = 0
	game.Num_right = 0
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
}

//献祭房间
type Russian struct {
	GameId     int
	Card_ID    string
	Time_start string
	Time_end   string
	LoginInfo  *LoginInfo
}

func NewRussian() *Russian {
	game := Russian{}
	game.GameId = ID_Russian
	game.LoginInfo = &LoginInfo{}
	game.LoginInfo.PlayerCardInfo = make(map[string]string)
	game.LoginInfo.CardTicketInfo = make(map[string]string)
	return &game
}

func (game *Russian) Reset() {
	game.Card_ID = ""
	game.Time_start = ""
	game.Time_end = ""
	game.LoginInfo.PlayerNum = 0
	for k, _ := range game.LoginInfo.PlayerCardInfo {
		delete(game.LoginInfo.PlayerCardInfo, k)
	}
	for k, _ := range game.LoginInfo.CardTicketInfo {
		delete(game.LoginInfo.CardTicketInfo, k)
	}
}
