package core

const (
	api                       = "http://172.16.10.7/gsaleapi/"
	AuthorityGet              = api + "authority_list.php"
	GameDataAdivinacionCreate = api + "gamedata_adivinacion.php"
	GameDataAdivinacionModify = api + "14.php"
	GameDataBangCreate        = api + "gamedata_bang.php"
	GameDataBangModify        = api + "15.php"
	GameDataFollowCreate      = api + "gamedata_follow.php"
	GameDataFollowModify      = api + "16.php"
	GameDataGreetingCreate    = api + "gamedata_greeting.php"
	GameDataGreetingModify    = api + "17.php"
	GameDataHighnoonCreate    = api + "gamedata_highnoon.php"
	GameDataHighnoonModify    = api + "18.php"
	GameDataHunterCreate      = api + "gamedata_hunter.php"
	GameDataHunterModify      = api + "19.php"
	GameDataHunterBoxCreate   = api + "gamedata_hunter_box.php"
	GameDataHunterBoxModify   = api + "20.php"
	GameDataMarksmanCreate    = api + "gamedata_marksman.php"
	GameDataMarksmanModify    = api + "21.php"
	GameDataMinerCreate       = api + "gamedata_miner.php"
	GameDataMinerModify       = api + "22.php"
	GameDataPrivityCreate     = api + "gamedata_privity.php"
	GameDataPrivityModify     = api + "23.php"
	GameDataRussianCreate     = api + "gamedata_russian.php"
	GameDataRussianModify     = api + "24.php"
	TicketUse                 = api + "ticket_update.php"
	TicketCheck               = api + "ticket_game.php"
)

type HttpResponse struct {
	Data     string
	JsonData map[string]interface{}
	Api      string
	Msg      *InboxMessage
	//ArduinoId string //该消息收到后应该反馈的arduinoID
	//CardId    string //请求该消息的CardId
}

func NewHttpResponse() *HttpResponse {
	res := HttpResponse{}
	res.JsonData = make(map[string]interface{})
	return &res
}

func (res *HttpResponse) Get(key string) interface{} {
	if v, ok := res.JsonData[key]; ok {
		return v
	}
	return nil
}

func (res *HttpResponse) Set(key string, value interface{}) {
	res.JsonData[key] = value
}

func (res *HttpResponse) GetStr(key string) string {
	if v, ok := res.JsonData[key]; ok {
		return v.(string)
	}
	return ""
}
