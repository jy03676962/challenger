package core

import (
	"log"
	"time"
)

var _ = log.Printf

type Player struct {
	Pos            RP      `json:"pos"`
	Direction      string  `json:"dir"` // values:up,right,down,left
	Button         string  `json:"button"`
	ButtonTime     float64 `json:"buttonTime"`
	ButtonLevel    int     `json:"buttonLevel"`
	Gold           int     `json:"gold"`
	Energy         float64 `json:"energy"`
	LevelData      [4]int  `json:"levelData"`
	HitCount       int     `json:"hitCount"`
	LostGold       int     `json:"lostgold"`
	InvincibleTime float64 `json:"invincibleTime"`
	Combo          int     `json:"combo"`
	ComboCount     int     `json:"comboCount"`
	ControllerID   string  `json:"cid"`
	DisplayPos     RP      `json:"displayPos"`
	Offline        int     `json:"offline"`

	moving      bool
	lastButton  string
	lastHitTime time.Time
	isSimulator bool
	tilePos     P
	status      string
}

func NewPlayer(cid string, isSimulator bool) *Player {
	p := Player{}
	p.Pos = RP{0, 0}
	p.Direction = "up"
	p.ControllerID = cid
	p.LevelData = [4]int{0, 0, 0, 0}
	p.lastHitTime = time.Unix(0, 0)
	p.isSimulator = isSimulator
	p.status = ""
	p.Offline = 0
	return &p
}

func (p *Player) setOffline() {
	p.Offline = 1
	p.status = ""
}

func (p *Player) setOnline() {
	p.Offline = 0
}

func (p *Player) updateLoc(loc int) {
	opt := GetOptions()
	loc = opt.TransferWearableLocation(loc) - 1
	if tp, valid := opt.TryIntToTile(loc); valid {
		p.tilePos = tp
		p.Pos = opt.RealPosition(p.tilePos)
		y := loc / opt.ArenaWidth
		x := loc % opt.ArenaWidth
		y = opt.ArenaHeight - 1 - y
		p.DisplayPos = opt.RealPosition(P{x, y})
	}
}

func (p *Player) UpdatePos(sec float64, options *MatchOptions) bool {
	if !p.moving {
		return false
	}
	delta := sec * options.PlayerSpeed
	var dx, dy float64
	switch p.Direction {
	case "up":
		dx = 0
		dy = -delta
	case "right":
		dx = delta
		dy = 0
	case "down":
		dx = 0
		dy = delta
	case "left":
		dx = -delta
		dy = 0
	}
	minXY := (float64(options.ArenaBorder) + options.PlayerSize) / 2
	maxX := float64((options.ArenaBorder+options.ArenaCellSize)*options.ArenaWidth) - minXY
	maxY := float64((options.ArenaBorder+options.ArenaCellSize)*options.ArenaHeight) - minXY
	x := MinMaxfloat64(p.Pos.X+dx, minXY, maxX)
	y := MinMaxfloat64(p.Pos.Y+dy, minXY, maxY)
	size := float64(options.PlayerSize)
	rect := Rect{
		float64(x) - size/2,
		float64(y) - size/2,
		size,
		size,
	}
	if !options.CollideWall(&rect) {
		p.Pos = RP{x, y}
		return true
	}
	return false
}

func (p *Player) Stay(sec float64, options *MatchOptions, rampage bool) {
	if p.Button != "" {
		p.ButtonTime += sec
		t := p.ButtonTime
		level := 0
		if rampage {
			if t > options.TRampage {
				level = 1
			}
		} else {
			if t < options.T1 {
				level = 0
			} else if t < options.T2 {
				level = 1
			} else if t < options.T3 {
				level = 2
			} else {
				level = 3
			}
		}
		p.ButtonLevel = level
	} else {
		rect := Rect{
			float64(p.Pos.X) - float64(options.PlayerSize)/2,
			float64(p.Pos.Y) - float64(options.PlayerSize)/2,
			float64(options.PlayerSize),
			float64(options.PlayerSize),
		}
		buttons := options.PressingButtons(&rect)
		if buttons != nil {
			var id string
			if len(buttons[1]) > 0 {
				if buttons[0] == p.lastButton {
					id = buttons[1]
				} else {
					id = buttons[0]
				}
			} else {
				id = buttons[0]
			}
			p.Button = id
		}
	}
}
