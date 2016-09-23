package core

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

var _ = log.Printf

type MatchEventType int

const (
	MatchEventTypeEnd = iota
	MatchEventTypeUpdate
)

const (
	WarmupTriggerButtonNotStart = -1.0
)

type MatchEvent struct {
	Type MatchEventType
	ID   uint
	Data interface{}
}

type laserInfoChange struct {
	id  string
	idx int
}

type laserCommand struct {
	id   string
	idx  int
	isOn bool
}

type Match struct {
	Member         []*Player        `json:"member"`
	Stage          string           `json:"stage"`
	TotalTime      float64          `json:"totalTime"`
	Elasped        float64          `json:"elasped"`
	WarmupTime     float64          `json:"warmupTime"`
	RampageTime    float64          `json:"rampageTime"`
	Mode1MaxTime   float64          `json:"mode1MaxTime"`
	Mode           string           `json:"mode"`
	Gold           int              `json:"gold"`
	Energy         float64          `json:"energy"`
	OnButtons      map[string]bool  `json:"onButtons"`
	RampageCount   int              `json:"rampageCount"`
	Lasers         []LaserInterface `json:"lasers"`
	ID             uint             `json:"id"`
	TeamID         string           `json:"teamID"`
	MaxEnergy      float64          `json:"maxEnergy"`
	MaxRampageTime float64          `json:"maxRampageTime"`
	IsSimulator    int              `json:"isSimulator"`

	offButtons    []string
	hiddenButtons map[string]*float64
	goldDropTime  float64
	opt           *MatchOptions
	srv           *Srv
	msgCh         chan *InboxMessage
	closeCh       chan bool
	laserCmdCh    chan *laserCommand
	matchData     *MatchData
	isSimulator   bool
	laserStatus   map[int]bool
	syncCount     int
	receiverMap   map[string]bool
	// 热身阶段相关状态
	currentWarmupStage        int
	warmupTriggerButtonRemain float64
	warmupCellButtonStatus    []bool
}

func NewMatch(s *Srv, controllerIDs []string, matchData *MatchData, mode string, teamID string, isSimulator bool) *Match {
	m := Match{}
	m.srv = s
	m.Member = make([]*Player, len(controllerIDs))
	for i, id := range controllerIDs {
		m.Member[i] = NewPlayer(id, isSimulator)
	}
	m.ID = matchData.ID
	m.matchData = matchData
	m.Stage = "before"
	m.opt = GetOptions()
	m.Mode1MaxTime = m.opt.Mode1TotalTime
	m.Mode = mode
	m.receiverMap = GetLaserPair().GetValidReceivers(false)
	m.msgCh = make(chan *InboxMessage, 1000)
	m.closeCh = make(chan bool)
	m.TeamID = teamID
	m.MaxEnergy = GetOptions().MaxEnergy
	m.MaxRampageTime = m.opt.RampageTime[m.modeIndex()]
	m.laserCmdCh = make(chan *laserCommand)
	m.isSimulator = isSimulator
	m.warmupTriggerButtonRemain = WarmupTriggerButtonNotStart
	m.warmupCellButtonStatus = make([]bool, m.opt.ArenaWidth*m.opt.ArenaHeight)
	if isSimulator {
		m.IsSimulator = 1
	} else {
		m.IsSimulator = 0
	}
	m.syncCount = 0
	return &m
}

func (m *Match) Run() {
	dt := 33 * time.Millisecond
	tickChan := time.Tick(dt)
	if m.Mode == "g" {
		m.TotalTime = m.opt.Mode1TotalTime
	} else {
		m.Gold = m.opt.Mode2InitGold[len(m.Member)-1]
	}
	if m.isSimulator {
		for _, member := range m.Member {
			member.Pos = m.opt.RealPosition(m.opt.ArenaEntrance)
		}
	}
	m.WarmupTime = m.opt.Warmup
	m.setStage("warmup")
	if !m.isSimulator {
		go m.handleLaserCmd()
	}
	for {
		<-tickChan
		m.handleInputs()
		if m.Stage == "after" || m.Stage == "stop" {
			break
		}
		m.tick(dt)
		m.sync()
	}
	d := make(map[string]interface{})
	d["matchData"] = m.dumpMatchData()
	d["teamID"] = m.TeamID
	m.srv.onMatchEvent(MatchEvent{MatchEventTypeEnd, m.ID, d})
	close(m.closeCh)
}

func (m *Match) OnMatchCmdArrived(cmd *InboxMessage) {
	go func() {
		select {
		case m.msgCh <- cmd:
		case <-m.closeCh:
		}
	}()
}

func (m *Match) OnLaserInfoArrived(msg *InboxMessage) {
	if m.isSimulator || GetOptions().CatchMode == 0 {
		return
	}
	m.OnMatchCmdArrived(msg)
}

func (m *Match) handleLaserCmd() {
	dict := make(map[string]*int)
	for {
		select {
		case cmd := <-m.laserCmdCh:
			key := cmd.id + ":" + strconv.Itoa(cmd.idx)
			v, ok := dict[key]
			sendCmd := false
			if cmd.isOn {
				if ok {
					sendCmd = *v == 0
					*v += 1
				} else {
					x := 1
					dict[key] = &x
					sendCmd = true
				}
			} else {
				if ok {
					*v -= 1
					sendCmd = *v == 0
					if *v < 0 {
						log.Println("warning:laser count less than 0:%v\n", *v)
						*v = 0
					}
				} else {
					log.Println("warning:laser count don't match")
				}
			}
			if sendCmd {
				m.srv.laserControl(cmd.id, cmd.idx, cmd.isOn)
			}
		case <-m.closeCh:
			return
		}
	}
}

func (m *Match) tick(dt time.Duration) {
	sec := dt.Seconds()
	if m.isWarmup() {
		m.WarmupTime = math.Max(m.WarmupTime-sec, 0)
		if m.warmupTriggerButtonRemain != WarmupTriggerButtonNotStart {
			m.warmupTriggerButtonRemain -= sec * 1000
			if m.warmupTriggerButtonRemain <= 0 {
				toOpenTiles := make(map[int]bool)
				for k, v := range m.warmupCellButtonStatus {
					if !v {
						continue
					}
					for _, i := range m.opt.TileAdjacency[m.opt.Conv(k)] {
						toOpenTiles[m.opt.Conv(i)] = true
					}
				}
				for k, _ := range toOpenTiles {
					if v := m.warmupCellButtonStatus[k]; v {
						continue
					}
					m.warmupCellButtonStatus[k] = true
					m.buttonControl(k, true)
				}
				m.warmupTriggerButtonRemain = m.opt.WarmupButtonInterval
			}
		}
		if m.currentWarmupStage < len(m.opt.WarmupLasers) {
			warmupLaser := m.opt.WarmupLasers[m.currentWarmupStage]
			elasped := m.opt.Warmup - m.WarmupTime
			if elasped*1000 >= float64(warmupLaser.Time) {
				if m.currentWarmupStage == 0 {
					m.warmupTriggerButtonRemain = m.opt.WarmupButtonInterval
					for _, player := range m.Member {
						p := m.opt.TilePosToInt(player.tilePos)
						m.warmupCellButtonStatus[p] = true
						m.buttonControl(p, true)
					}
				}
				m.srv.lasersControl(warmupLaser.Large[:], warmupLaser.Small[:])
				m.currentWarmupStage += 1
				if m.currentWarmupStage >= len(m.opt.WarmupLasers) {
					log.Println("stop warmup effect")
					m.warmupTriggerButtonRemain = WarmupTriggerButtonNotStart
					m.srv.setWallM2M3Auto(true)
				}
			}
		}
	} else if m.isOngoing() {
		m.Elasped += sec
		if m.Mode == "g" {
			m.TotalTime = math.Max(m.TotalTime-sec, 0)
		}
		m.RampageTime = math.Max(m.RampageTime-sec, 0)
		if m.Mode == "s" && m.goldDropTime > 0 && m.RampageTime <= 0 {
			m.goldDropTime -= sec
			if m.goldDropTime <= 0 {
				m.Gold -= m.opt.Mode2GoldDropRate[len(m.Member)-1]
				m.goldDropTime = m.opt.Mode2GoldDropInterval
			}
		}
		for k, v := range m.hiddenButtons {
			*v -= sec
			if *v <= 0 {
				delete(m.hiddenButtons, k)
				m.OnButtons[k] = true
				m.setSingleButtonEffect(k)
			}
		}
	}
	for _, player := range m.Member {
		m.playerTick(player, sec)
	}
	for _, laser := range m.Lasers {
		laser.Tick(sec)
	}
	m.updateStage()
}

func (m *Match) setStage(s string) {
	if m.Stage == s {
		return
	}
	switch s {
	case "warmup":
		m.srv.bgControl(m.opt.BgWarmup[m.modeIndex()])
		m.srv.ledControl(3, "0", "1", "2", "3")
		if m.Mode == "g" {
			m.srv.doorControl("5", "5", "D-1")
			m.srv.doorControl("5", "5", "D-2")
			m.srv.doorControl("", "5", "D-3")
			m.srv.doorControl("", "5", "D-4")
		} else {
			m.srv.doorControl("12", "12", "D-1")
			m.srv.doorControl("12", "12", "D-2")
			m.srv.doorControl("", "12", "D-3")
			m.srv.doorControl("", "12", "D-4")
		}
	case "ongoing-low-0":
		if m.Mode == "g" {
			m.srv.doorControl("5", "5", "D-1")
			m.srv.doorControl("5", "5", "D-2")
			m.srv.doorControl("", "5", "D-3")
			m.srv.doorControl("", "5", "D-4")
		} else {
			m.srv.doorControl("12", "12", "D-1")
			m.srv.doorControl("12", "12", "D-2")
			m.srv.doorControl("", "12", "D-3")
			m.srv.doorControl("", "12", "D-4")
		}
		m.srv.lightControl("1")
		if m.Stage == "ongoing-rampage" {
			msg := NewInboxMessage()
			msg.SetCmd("btn_ctrl")
			msg.Set("useful", "0")
			if m.Mode == "g" {
				msg.Set("mode", "1")
			} else {
				msg.Set("mode", "2")
			}
			msg.Set("stage", "0")
			m.srv.sends(msg, InboxAddressTypeMainArduinoDevice)
			m.initButtons()
		} else if m.isWarmup() {
			if m.Mode == "s" {
				m.goldDropTime = m.opt.Mode2GoldDropInterval
			}
			m.initLasers()
			m.initButtons()
		}
		m.srv.bgControl(m.opt.BgNormal[m.modeIndex()])
		if m.Mode == "g" {
			m.srv.ledControl(3, "5")
			m.srv.ledControl(1, "0", "2", "3")
		} else {
			m.srv.ledControl(3, "12")
			m.srv.ledControl(1, "0", "2", "3")
		}
	case "ongoing-low-1":
	case "ongoing-low-2":
	case "ongoing-low-3":
		level, _ := strconv.Atoi(strings.Split(s, "-")[2])
		var mode string
		if m.Mode == "g" {
			mode = strconv.Itoa(level + 5)
		} else {
			mode = strconv.Itoa(level + 12)
		}
		m.srv.ledControl(3, mode)
	case "ongoing-high":
		m.srv.lightControl("2")
		m.srv.bgControl(m.opt.BgHigh[m.modeIndex()])
		if m.Mode == "g" {
			m.srv.doorControl("9", "", "D-1")
			m.srv.doorControl("9", "", "D-2")
			m.srv.ledControl(3, "9")
		} else {
			m.srv.doorControl("16", "", "D-1")
			m.srv.doorControl("16", "", "D-2")
			m.srv.ledControl(3, "16")
		}
	case "ongoing-full":
		m.srv.bgControl(m.opt.BgFull[m.modeIndex()])
		if m.Mode == "g" {
			m.srv.ledControl(3, "19")
		} else {
			m.srv.ledControl(3, "20")
		}
	case "ongoing-rampage":
		m.srv.bgControl(m.opt.BgRampage[m.modeIndex()])
		m.srv.lightControl("0")
		m.srv.doorControl("42", "", "D-1")
		m.srv.doorControl("42", "", "D-2")
		m.RampageTime = m.opt.RampageTime[m.modeIndex()]
		laserPosList := make([]int, len(m.Lasers))
		offButtons := make(map[string]bool)
		for i, laser := range m.Lasers {
			laserPosList[i] = laser.Pause(m.RampageTime)
		}
		for _, p := range laserPosList {
			infos := m.opt.mainArduinoInfosByPos(p)
			for _, info := range infos {
				offButtons[info.ID] = true
			}
		}
		m.offButtons = make([]string, len(offButtons))
		m.hiddenButtons = make(map[string]*float64)
		addrs := make([]InboxAddress, len(offButtons))
		offIdx := 0
		for _, btn := range m.opt.Buttons {
			if _, ok := offButtons[btn.Id]; !ok {
				m.OnButtons[btn.Id] = true
			} else {
				m.offButtons[offIdx] = btn.Id
				addrs[offIdx] = InboxAddress{InboxAddressTypeMainArduinoDevice, btn.Id}
				offIdx += 1
			}
		}
		m.setButtonEffect("2", false)
		m.Energy = 0
		m.RampageCount += 1
		for _, player := range m.Member {
			player.Combo = 0
			player.lastHitTime = time.Unix(0, 0)
		}
		m.srv.ledRampageEffect(offButtons)
	case "ongoing-countdown":
		m.srv.bgControl(m.opt.BgCountdown[m.modeIndex()])
		m.srv.ledControl(1, "47")
		m.srv.ledControl(2, "46")
	case "after", "stop":
		m.srv.bgControl(m.opt.BgLeave[m.modeIndex()])
		m.srv.doorControl("23", "1", "D-1")
		m.srv.doorControl("46", "1", "D-2")
		m.srv.doorControl("", "1", "D-3")
		m.srv.doorControl("", "1", "D-4")
		for _, laser := range m.Lasers {
			laser.Close()
		}
		for _, player := range m.Member {
			m.updatePlayerStatus("01", player)
		}
		m.srv.setWallM2M3Auto(false)
		m.srv.ledFlowEffect()
	}
	log.Printf("game stage:%v\n", s)
	m.Stage = s
}

func (m *Match) updateStage() {
	if m.RampageTime > 0 {
		m.setStage("ongoing-rampage")
		return
	}
	if m.WarmupTime > 0 {
		m.setStage("warmup")
		return
	}
	s := m.Stage
	level := int(m.Energy/m.opt.MaxEnergy*100) / 20
	if level < 4 {
		s = fmt.Sprintf("ongoing-low-%d", level)
	} else if level < 5 {
		s = "ongoing-high"
	} else {
		if len(m.Member) == 1 {
			s = "ongoing-rampage"
		} else {
			together := true
			if m.isSimulator {
				p, pBool := m.opt.TilePosition(m.Member[0].Pos)
				if pBool {
					for i := 1; i < len(m.Member); i++ {
						pp, ppBool := m.opt.TilePosition(m.Member[i].Pos)
						if !ppBool || pp.X != p.X || pp.Y != p.Y {
							together = false
							break
						}
					}
				} else {
					together = false
				}
			} else {
				tp := m.Member[0].tilePos
				for i := 1; i < len(m.Member); i++ {
					if m.Member[i].tilePos.X != tp.X || m.Member[i].tilePos.Y != tp.Y {
						together = false
					}
				}
			}
			if together {
				s = "ongoing-rampage"
			} else {
				s = "ongoing-full"
			}
		}
	}
	if m.Mode == "g" && s != "ongoing-rampage" && m.TotalTime < m.opt.Mode1CountDown {
		s = "ongoing-countdown"
	}
	if m.Mode == "g" && m.TotalTime <= 0 || m.Mode == "s" && m.Gold <= 0 {
		s = "after"
	}
	m.setStage(s)
}

func (m *Match) openLaser(ID string, idx int) {
	m.laserCmdCh <- &laserCommand{ID, idx, true}
}

func (m *Match) closeLaser(ID string, idx int) {
	m.laserCmdCh <- &laserCommand{ID, idx, false}
}

func (m *Match) sync() {
	m.syncCount += 1
	if m.isSimulator {
		b, _ := json.Marshal(m)
		m.srv.onMatchEvent(MatchEvent{MatchEventTypeUpdate, m.ID, string(b)})
	} else if m.syncCount >= 30 {
		b, _ := json.Marshal(m)
		m.srv.onMatchEvent(MatchEvent{MatchEventTypeUpdate, m.ID, string(b)})
		m.syncCount = 0
	}
}

func (m *Match) reset() {
	m.Member = make([]*Player, 0)
	m.Stage = "before"
	m.TotalTime = 0
	m.Elasped = 0
	m.WarmupTime = 0
	m.RampageTime = 0
	m.Mode = ""
	m.Gold = 0
	m.Energy = 0
	m.OnButtons = nil
	m.RampageCount = 0
	m.Lasers = nil
	m.offButtons = nil
	m.hiddenButtons = nil
	m.goldDropTime = 0
}

func (m *Match) handleInputs() bool {
	hasInputs := false
	for {
		select {
		case msg := <-m.msgCh:
			hasInputs = true
			m.handleInput(msg)
		default:
			return hasInputs
		}
	}
}

func (m *Match) handleInput(msg *InboxMessage) {
	if msg.RemoveAddress != nil {
		m.playerOffline(msg.RemoveAddress.String())
		return
	}
	if msg.AddAddress != nil {
		m.playerOnline(msg.AddAddress.String())
		return
	}
	cmd := msg.GetCmd()
	switch cmd {
	case "stopMatch":
		m.setStage("stop")
	case "playerMove":
		if player := m.getPlayer(msg.Address.String()); player != nil {
			player.moving = true
			player.Direction = msg.GetStr("dir")
		}
	case "playerStop":
		if player := m.getPlayer(msg.Address.String()); player != nil {
			player.moving = false
		}
	case "wearableLoc":
		if player := m.getPlayer(msg.Address.String()); player != nil {
			if status := msg.GetStr("status"); len(status) > 0 {
				player.status = status
			}
			loc, _ := strconv.Atoi(msg.GetStr("loc"))
			if loc > 0 {
				player.updateLoc(loc)
			}
		}
	case "upload_score":
		if !m.isOngoing() {
			break
		}
		info := arduinoInfoFromID(msg.Address.ID)
		for _, player := range m.Member {
			if player.tilePos.X == info.X-1 && player.tilePos.Y == info.Y-1 {
				m.consumeButton(info.ID, player, msg.GetStr("score"))
				break
			}
		}
		m.onButtonPressed(info.ID)
	case "hb":
		ur := msg.GetStr("UR")
		id := msg.GetStr("ID")
		changed := false
		for i, r := range ur {
			c := string(r)
			var isOn = false
			if c == "1" {
				isOn = true
			}
			key := id + ":" + strconv.Itoa(i)
			if old, ok := m.receiverMap[key]; ok {
				m.receiverMap[key] = isOn
				if old != isOn {
					changed = true
				}
			}
		}
		if changed {
			musicPostions := make(map[int]bool)
			for _, laser := range m.Lasers {
				l := laser.(*Laser)
				blocked, p, _ := l.IsTouched(m.receiverMap)
				if blocked {
					shouldPause := false
					for _, player := range m.Member {
						pp := GetOptions().TilePosToInt(player.tilePos)
						if pp == p && player.InvincibleTime <= 0 {
							musicPostions[pp] = true
							m.touchPunish(player)
							shouldPause = true
						}
					}
					if shouldPause {
						l.Pause(GetOptions().LaserPauseTime)
					}
				}
			}
			for pos, _ := range musicPostions {
				tilePos := GetOptions().IntToTile(pos)
				m.srv.musicControlByCell(tilePos.X, tilePos.Y, "6")
			}
		}
	}
}

func (m *Match) getPlayer(controllerID string) *Player {
	for _, player := range m.Member {
		if player.ControllerID == controllerID {
			return player
		}
	}
	return nil
}

func (m *Match) playerOffline(cid string) {
	for _, player := range m.Member {
		if player.ControllerID == cid {
			player.setOffline()
		}
	}
}

func (m *Match) playerOnline(cid string) {
	for _, player := range m.Member {
		if player.ControllerID == cid {
			player.setOnline()
		}
	}
}

func (m *Match) playerTick(player *Player, sec float64) {
	player.InvincibleTime = math.Max(player.InvincibleTime-sec, 0)
	if m.isSimulator {
		moved := player.UpdatePos(sec, m.opt)
		if !m.isOngoing() {
			return
		}
		if moved && player.Button != "" {
			btn := player.Button
			m.consumeButton(player.Button, player, "")
			m.onButtonPressed(btn)
		}
		if !moved {
			player.Stay(sec, m.opt, m.RampageTime > 0)
		}
	} else {
		if player.InvincibleTime > 0 {
			m.updatePlayerStatus("05", player)
		} else {
			if m.Stage == "ongoing-full" {
				m.updatePlayerStatus("03", player)
			} else if m.Stage == "ongoing-rampage" {
				m.updatePlayerStatus("04", player)
			} else {
				m.updatePlayerStatus("02", player)
			}
		}
	}
}

func (m *Match) updatePlayerStatus(st string, p *Player) {
	if p.status != st {
		m.srv.wearableControl(st, p.ControllerID)
	}
}

func (m *Match) dumpMatchData() *MatchData {
	m.matchData.Mode = m.Mode
	m.matchData.Elasped = m.Elasped
	m.matchData.Member = make([]PlayerData, 0)
	m.matchData.RampageCount = m.RampageCount
	m.matchData.AnswerType = MatchNotAnswer
	m.matchData.TeamID = m.TeamID
	m.matchData.ExternalID = ""
	totalGold := 0
	m.matchData.Grade = m.opt.TeamGrade(m.Gold, m.Elasped, len(m.Member), m.Mode)
	for _, player := range m.Member {
		totalGold += player.Gold - player.LostGold
		playerData := PlayerData{}
		playerData.Gold = player.Gold
		playerData.LostGold = player.LostGold
		playerData.Energy = player.Energy
		playerData.Combo = player.ComboCount
		strs := make([]string, 4)
		for i, c := range player.LevelData {
			strs[i] = strconv.Itoa(c)
		}
		playerData.LevelData = strings.Join(strs, ",")
		playerData.HitCount = player.HitCount
		playerData.Name = ""
		playerData.QuestionInfo = ""
		playerData.Answered = 0
		playerData.ExternalID = ""
		playerData.ControllerID = player.ControllerID
		playerData.Grade = m.opt.PersonGrade(player.Gold-player.LostGold, len(m.Member), m.Mode)
		m.matchData.Member = append(m.matchData.Member, playerData)
	}
	m.matchData.Gold = totalGold
	return m.matchData
}

func (m *Match) modeIndex() int {
	if m.Mode == "g" {
		return 0
	} else {
		return 1
	}
}

func (m *Match) initLasers() {
	m.Lasers = make([]LaserInterface, len(m.Member))
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	l := r.Perm(m.opt.ArenaWidth * m.opt.ArenaHeight)
	for i, player := range m.Member {
		loc := l[i]
		p := P{loc % m.opt.ArenaWidth, loc / m.opt.ArenaWidth}
		if m.isSimulator {
			m.Lasers[i] = NewSimuLaser(p, player, m)
		} else {
			m.Lasers[i] = NewLaser(p, player, m)
		}
	}
}

func (m *Match) touchPunish(p *Player) {
	opt := GetOptions()
	p.InvincibleTime = opt.PlayerInvincibleTime
	p.HitCount += 1
	var punish int
	playerCount := len(m.Member)
	if m.Mode == "g" {
		punish = opt.Mode1TouchPunish[playerCount-1]
	} else {
		punish = opt.Mode2TouchPunish[playerCount-1]
	}
	m.Gold = m.Gold - punish
	p.LostGold += punish
}

func (m *Match) initButtons() {
	for _, player := range m.Member {
		player.Button = ""
		player.lastButton = ""
		player.ButtonLevel = 0
		player.ButtonTime = 0
	}
	count := len(m.opt.Buttons)
	src := rand.NewSource(time.Now().UnixNano())
	r := rand.New(src)
	randList := r.Perm(count)
	n := m.opt.InitButtonNum[len(m.Member)-1]
	m.OnButtons = make(map[string]bool)
	m.offButtons = make([]string, count-n)
	m.hiddenButtons = make(map[string]*float64)
	for i, j := range randList {
		id := m.opt.Buttons[j].Id
		if i < n {
			m.OnButtons[id] = true
		} else {
			m.offButtons[i-n] = id
		}
	}
	if !m.isSimulator {
		m.setButtonEffect("0", true)
	}

}

func (m *Match) buttonControl(p int, isOn bool) {
	arduinos := m.opt.mainArduinoInfosByPos(p)
	addrs := make([]InboxAddress, len(arduinos))
	for i, info := range arduinos {
		addrs[i] = InboxAddress{InboxAddressTypeMainArduinoDevice, info.ID}
	}
	msg := NewInboxMessage()
	msg.SetCmd("btn_ctrl")
	if isOn {
		msg.Set("useful", "3")
	} else {
		msg.Set("useful", "0")
	}
	if m.Mode == "g" {
		msg.Set("mode", "1")
	} else {
		msg.Set("mode", "2")
	}
	msg.Set("stage", "0")
	m.srv.send(msg, addrs)
}

func (m *Match) setButtonEffect(stage string, immediately bool) {
	onAddrs := make([]InboxAddress, len(m.OnButtons))
	offAddrs := make([]InboxAddress, len(m.offButtons))
	i := 0
	for id, _ := range m.OnButtons {
		onAddrs[i] = InboxAddress{InboxAddressTypeMainArduinoDevice, id}
		i += 1
	}
	for i, id := range m.offButtons {
		offAddrs[i] = InboxAddress{InboxAddressTypeMainArduinoDevice, id}
	}
	if len(onAddrs) > 0 {
		msg := NewInboxMessage()
		msg.SetCmd("btn_ctrl")
		if immediately {
			msg.Set("useful", "3")
		} else {
			msg.Set("useful", "1")
		}
		if m.Mode == "g" {
			msg.Set("mode", "1")
		} else {
			msg.Set("mode", "2")
		}
		msg.Set("stage", stage)
		m.srv.send(msg, onAddrs)
	}
	if len(offAddrs) > 0 {
		msg := NewInboxMessage()
		msg.SetCmd("btn_ctrl")
		msg.Set("useful", "0")
		if m.Mode == "g" {
			msg.Set("mode", "1")
		} else {
			msg.Set("mode", "2")
		}
		msg.Set("stage", stage)
		m.srv.send(msg, offAddrs)
	}
}

func (m *Match) setSingleButtonEffect(id string) {
	if !m.isOngoing() {
		return
	}
	msg := NewInboxMessage()
	msg.SetCmd("btn_ctrl")
	msg.Set("useful", "1")
	if m.Mode == "g" {
		msg.Set("mode", "1")
	} else {
		msg.Set("mode", "2")
	}
	if strings.HasPrefix(m.Stage, "ongoing-low") {
		msg.Set("stage", "0")
	} else {
		msg.Set("stage", "1")
	}
	addr := InboxAddress{InboxAddressTypeMainArduinoDevice, id}
	m.srv.sendToOne(msg, addr)
}

func (m *Match) consumeButton(btn string, player *Player, lvl string) {
	level := 0
	if !m.isSimulator {
		switch lvl {
		case "S":
			level = 1
		case "A":
			level = 2
		case "B":
			level = 3
		case "M":
			level = 0
		}
	} else {
		level = player.ButtonLevel
	}
	player.LevelData[level] += 1
	if level > 0 {
		bonus := m.opt.GoldBonus[m.modeIndex()]
		m.Gold += bonus
		player.Gold += bonus
		if m.RampageTime <= 0 {
			sec := time.Since(player.lastHitTime).Seconds()
			player.lastHitTime = time.Now()
			var max float64
			if player.Combo == 0 {
				max = m.opt.FirstComboInterval[len(m.Member)-1]
			} else {
				max = m.opt.FirstComboInterval[len(m.Member)-1]
			}
			if sec <= max {
				player.Combo += 1
			} else {
				player.Combo = 0
			}
			extra := 0.0
			if player.Combo == 1 {
				extra = m.opt.FirstComboExtra
				player.ComboCount += 1
			} else if player.Combo > 1 {
				extra = m.opt.ComboExtra
			}
			delta := m.opt.EnergyBonus[level][len(m.Member)-1] + extra
			m.Energy = math.Min(m.opt.MaxEnergy, m.Energy+delta)
			player.Energy += delta
		}
	}
	player.lastButton = btn
	player.ButtonLevel = 0
	player.Button = ""
	player.ButtonTime = 0
}

func (m *Match) onButtonPressed(btn string) {
	if m.RampageTime <= 0 {
		delete(m.OnButtons, btn)
		src := rand.NewSource(time.Now().UnixNano())
		r := rand.New(src)
		i := r.Intn(len(m.offButtons))
		key := m.offButtons[i]
		m.offButtons[i] = btn
		t := m.opt.ButtonHideTime[m.modeIndex()]
		m.hiddenButtons[key] = &t
	}
}

func (m *Match) musicControlByCell(x int, y int, music string) {
	m.srv.musicControlByCell(x, y, music)
}

func (m *Match) isWarmup() bool {
	return strings.HasPrefix(m.Stage, "warmup")
}

func (m *Match) isOngoing() bool {
	return strings.HasPrefix(m.Stage, "ongoing")
}
