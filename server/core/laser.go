package core

import (
	"log"
	"math"
	"strconv"
)

const laserSize = 10

var _ = log.Printf

var catchByPos = false

type LaserLine struct {
	ID      string
	Index   int
	P       int
	elasped float64
}

type Laser struct {
	IsPause              bool    `json:"isPause"`
	Warning              float64 `json:"warning"`
	DisplayP             RP      `json:"displayP"`
	DisplayP2            RP      `json:"displayP2"`
	player               *Player
	dest                 int
	pathMap              map[int]int
	p                    int
	p2                   int
	match                *Match
	pauseTime            float64
	elaspedSinceLastMove float64
	lines                []*LaserLine
	startupLines         []*LaserLine
	startupingIndex      int
	closed               bool
}

func NewLaser(p P, player *Player, match *Match) *Laser {
	l := Laser{}
	l.IsPause = true
	l.player = player
	l.dest = -1
	l.match = match
	l.pathMap = make(map[int]int)
	l.p = GetOptions().TilePosToInt(p)
	l.p2 = -1
	l.convertDisplay()
	l.elaspedSinceLastMove = GetOptions().LaserSpeed
	l.lines = make([]*LaserLine, 0)
	l.startupLines = l.linesByP(l.p)
	l.startupingIndex = 0
	l.closed = false
	l.Warning = GetOptions().LaserAppearTime
	match.musicControlByCell(p.X, p.Y, "4")
	match.srv.ledControlByCell(p.X, p.Y, "24")
	return &l
}

func (l *Laser) Pause(t float64) int {
	if l.isStartuping() {
		return l.p
	}
	l.IsPause = true
	l.pauseTime = math.Max(t, l.pauseTime)
	l.doClose()
	p, p2 := 0, 0
	for _, line := range l.lines {
		if line.P == l.p {
			p += 1
		} else {
			p2 += 1
		}
	}
	if p2 > p {
		l.p = l.p2
	}
	l.p2 = -1
	l.lines = l.linesByP(l.p)
	return l.p
}

func (l *Laser) IsFollow(cid string) bool {
	return l.player.ControllerID == cid
}

func (l *Laser) Close() {
	l.closed = true
	l.doClose()
}

func (l *Laser) IsTouched(m map[string]bool) (touched bool, p int, senderID string) {
	p = 0
	touched = false
	senderID = ""
	if l.IsPause || l.isStartuping() {
		return
	}
	for k, v := range m {
		if v {
			continue
		}
		for _, line := range l.lines {
			if line.elasped < 1000 {
				continue
			}
			info := GetLaserPair().Get(line.ID, line.Index)
			if info != nil && info.Valid > 0 && (info.ID+":"+info.Idx) == k {
				p = line.P
				senderID = line.ID + ":" + strconv.Itoa(line.Index)
				touched = true
				return
			}
		}
	}
	return
}

func (l *Laser) Tick(dt float64) {
	opt := GetOptions()
	if l.closed {
		return
	}
	if l.IsPause {
		l.pauseTime -= dt
		if l.pauseTime <= 0 {
			l.IsPause = false
			l.pauseTime = 0
			l.doOpen()
		}
		return
	}
	if l.Warning > 0 {
		l.Warning -= dt
		if l.Warning <= 0 {
			l.Warning = 0
			l.musicControlByPos(l.p, "5")
			tp := opt.IntToTile(l.p)
			l.match.srv.ledControlByCell(tp.X, tp.Y, "5")
		}
		return
	}
	for _, line := range l.lines {
		line.elasped += dt * 1000
	}
	l.elaspedSinceLastMove += dt
	interval := opt.laserMoveInterval(l.match.Energy, len(l.match.Member))
	if l.elaspedSinceLastMove < interval {
		return
	}
	l.elaspedSinceLastMove = 0
	if l.isStartuping() {
		line := l.startupLines[l.startupingIndex]
		l.match.openLaser(line.ID, line.Index)
		l.lines = append(l.lines, line)
		l.startupingIndex += 1
	} else {
		if l.player.Offline > 0 {
			return
		}
		next := l.findPath()
		if l.p2 < 0 && l.p == next {
			return
		}
		replaceIdx := -1
		notInNext := 0
		for i, line := range l.lines {
			if line.P != next {
				notInNext += 1
				if replaceIdx < 0 {
					replaceIdx = i
				}
			}
		}
		if replaceIdx >= 0 {
			infos := opt.mainArduinoInfosByPos(next)
			for _, info := range infos {
				for i := 0; i < info.LaserNum; i++ {
					if !l.contains(info.ID, i) {
						line := l.lines[replaceIdx]
						l.match.closeLaser(line.ID, line.Index)
						l.match.openLaser(info.ID, i)
						l.lines[replaceIdx] = &LaserLine{info.ID, i, next, 0}
						if notInNext == 1 {
							if l.p != next {
								l.musicControlByPos(l.p, "0")
							}
							l.p = next
							l.p2 = -1
						} else {
							if l.p2 != next {
								l.musicControlByPos(next, "5")
							}
							l.p2 = next
						}
						l.convertDisplay()
						return
					}
				}
			}
		}
	}
}

func (l *Laser) musicControlByPos(p int, music string) {
	if p < 0 {
		return
	}
	tp := opt.IntToTile(p)
	l.match.musicControlByCell(tp.X, tp.Y, music)
}

func (l *Laser) doClose() {
	for _, line := range l.lines {
		l.match.closeLaser(line.ID, line.Index)
	}
	l.musicControlByPos(l.p, "0")
	l.musicControlByPos(l.p2, "0")
}

func (l *Laser) doOpen() {
	for _, line := range l.lines {
		l.match.openLaser(line.ID, line.Index)
	}
}

func (l *Laser) linesByP(p int) []*LaserLine {
	infos := GetOptions().mainArduinoInfosByPos(p)
	ret := make([]*LaserLine, 0)
	for _, info := range infos {
		for i := 0; i < info.LaserNum; i++ {
			ret = append(ret, &LaserLine{info.ID, i, l.p, 0})
		}
	}
	return ret
}

func (l *Laser) convertDisplay() {
	opt := GetOptions()
	y := l.p / opt.ArenaWidth
	x := l.p % opt.ArenaWidth
	y = opt.ArenaHeight - 1 - y
	l.DisplayP = opt.RealPosition(P{x, y})
	if l.p2 >= 0 {
		y := l.p2 / opt.ArenaWidth
		x := l.p2 % opt.ArenaWidth
		y = opt.ArenaHeight - 1 - y
		l.DisplayP2 = opt.RealPosition(P{x, y})
	} else {
		l.DisplayP2 = RP{-1, -1}
	}
}

func (l *Laser) contains(id string, idx int) bool {
	for _, line := range l.lines {
		if line.ID == id && line.Index == idx {
			return true
		}
	}
	return false
}

func (l *Laser) isStartuping() bool {
	return l.startupingIndex < len(l.startupLines)
}

func (l *Laser) findPath() int {
	opt := GetOptions()
	l.fillPath()
	pp1 := opt.Conv(l.p)
	if l.p2 >= 0 {
		pp2 := opt.Conv(l.p2)
		if l.pathMap[pp2] <= l.pathMap[pp1] {
			return l.p2
		} else {
			return l.p
		}
	}
	next, min := pp1, l.pathMap[pp1]
	for _, i := range opt.TileAdjacency[pp1] {
		if l.pathMap[i] < min {
			min = l.pathMap[i]
			next = i
		}
	}
	return opt.Conv(next)
}

func (l *Laser) fillPath() {
	opt := GetOptions()
	dest := opt.TilePosToInt(l.player.tilePos)
	if l.dest == dest {
		return
	}
	l.dest = dest
	for i := 0; i < opt.ArenaWidth*opt.ArenaHeight; i++ {
		l.pathMap[i] = 10000
	}
	var fill func(x int, v int)
	fill = func(x int, v int) {
		l.pathMap[x] = v
		for _, i := range opt.TileAdjacency[x] {
			if l.pathMap[i] > v+1 {
				fill(i, v+1)
			}
		}
	}
	fill(opt.Conv(l.dest), 0)
}
