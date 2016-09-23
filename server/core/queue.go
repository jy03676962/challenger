package core

import (
	"container/list"
	"log"
	"strconv"
	"sync"
)

const (
	initCursor       = 2000
	singleWaitTime   = 360
	maxFinishedCount = 3
)

type TeamStatus int

const (
	TS_Waiting  TeamStatus = iota
	TS_Prepare  TeamStatus = iota
	TS_Playing  TeamStatus = iota
	TS_After    TeamStatus = iota
	TS_Finished TeamStatus = iota
)

var _ = log.Printf

type Team struct {
	Size       int        `json:"size"`
	ID         string     `json:"id"`
	DelayCount int        `json:"delayCount"`
	Status     TeamStatus `json:"status"`
	WaitTime   int        `json:"waitTime"`
	Mode       string     `json:"mode"`
	Calling    int        `json:"calling"`
}

type Queue struct {
	li   *list.List
	srv  *Srv
	dict map[string]*list.Element
	cur  int
	lock *sync.RWMutex
}

func NewQueue(srv *Srv) *Queue {
	q := Queue{}
	q.li = list.New()
	q.srv = srv
	q.dict = make(map[string]*list.Element)
	q.cur = initCursor
	q.lock = new(sync.RWMutex)
	return &q
}

func (q *Queue) AddTeamToQueue(teamSize int, mode string) int {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	q.cur += 1
	id := strconv.Itoa(q.cur)
	t := Team{Size: teamSize, ID: id, Status: TS_Waiting, Mode: mode, Calling: 0}
	element := q.li.PushBack(&t)
	q.dict[id] = element
	return q.cur
}

func (q *Queue) ResetQueue() int {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	q.li.Init()
	q.dict = make(map[string]*list.Element)
	q.cur = initCursor
	return q.cur
}

func (q *Queue) TeamPrepare(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	if team.Status == TS_Waiting {
		team.Status = TS_Prepare
		team.Calling = 0
	}
}

func (q *Queue) TeamCancelPrepare(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	if team.Status == TS_Prepare {
		team.Status = TS_Waiting
		team.Calling = 0
	}
}

func (q *Queue) TeamStart(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	if team.Status == TS_Prepare {
		team.Status = TS_Playing
	}
}

func (q *Queue) TeamFinishMatch(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	q.li.Remove(element)
	delete(q.dict, teamID)
}

func (q *Queue) TeamCall(teamID string) {
	q.lock.RLock()
	defer q.lock.RUnlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	if team.Status == TS_Waiting {
		team.Calling = 1
	}
}

func (q *Queue) TeamCutLine(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	if team.Status != TS_Waiting {
		return
	}
	for e := q.li.Front(); e != nil; e = e.Next() {
		t := e.Value.(*Team)
		if t.Status == TS_Waiting && t.Calling == 0 {
			q.li.MoveBefore(element, e)
			return
		}
	}
}

func (q *Queue) TeamRemove(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	if team.Status != TS_Waiting {
		return
	}
	delete(q.dict, teamID)
	q.li.Remove(element)
}

func (q *Queue) TeamChangeMode(teamID string, mode string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	team.Mode = mode
}

func (q *Queue) TeamDelay(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	team.DelayCount += 1
	team.Calling = 0
	if team.DelayCount >= 4 {
		team.DelayCount = 4
		q.li.MoveToBack(element)
	} else {
		next := element.Next()
		if next != nil {
			q.li.MoveAfter(element, next)
		}
	}
}

func (q *Queue) TeamAddPlayer(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	team.Size += 1
}

func (q *Queue) TeamRemovePlayer(teamID string) {
	q.lock.Lock()
	defer q.lock.Unlock()
	defer q.updateHallData()
	element := q.dict[teamID]
	if element == nil {
		return
	}
	team := element.Value.(*Team)
	team.Size -= 1
}

func (q *Queue) TeamQueryData() {
	q.lock.RLock()
	q.updateHallData()
	q.lock.RUnlock()
}

func (q *Queue) GetAllTeamsFromQueue() []Team {
	result := make([]Team, q.li.Len())
	waitTime := 0
	for e, i := q.li.Front(), 0; e != nil; e, i = e.Next(), i+1 {
		team := e.Value.(*Team)
		if team.Status == TS_Waiting {
			waitTime += singleWaitTime
			team.WaitTime = waitTime
		}
		result[i] = *team
	}
	return result
}

func (q *Queue) updateHallData() {
	q.srv.onQueueUpdated(q.GetAllTeamsFromQueue())
}
