package core

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"log"
	"net/url"
	"time"
)

var _ = log.Printf

type MatchAnswerType int

const (
	MatchNotAnswer MatchAnswerType = 0
	MatchAnswering MatchAnswerType = 1
	MatchAnswered  MatchAnswerType = 2
)

type PlayerData struct {
	ID           uint      `json:"id"`
	CreatedAt    time.Time `json:"createdAt"`
	MatchID      int       `json:"-"`
	ExternalID   string    `gorm:"index" json:"eid"`
	Name         string    `json:"name"`
	Gold         int       `json:"gold"`
	LostGold     int       `json:"lostGold"`
	Energy       float64   `json:"energy"`
	Combo        int       `json:"combo"`
	Grade        string    `json:"grade"`
	Level        int       `json:"level"`
	LevelData    string    `json:"levelData"`
	HitCount     int       `json:"hitCount"`
	ControllerID string    `json:"cid"`
	QuestionInfo string    `json:"questionInfo"`
	Answered     int       `json:"answered"`
}

func (PlayerData) TableName() string {
	return "players"
}

type MatchData struct {
	ID           uint            `json:"id"`
	CreatedAt    time.Time       `json:"createdAt"`
	Mode         string          `json:"mode"`
	Elasped      float64         `json:"elasped"`
	Gold         int             `json:"gold"`
	Member       []PlayerData    `gorm:"ForeignKey:MatchID" json:"member"`
	RampageCount int             `json:"rampageCount"`
	AnswerType   MatchAnswerType `json:"answerType"`
	TeamID       string          `json:"teamID"`
	ExternalID   string          `gorm:"index" json:"eid"`
	Grade        string          `json:"grade"`
}

func (MatchData) TableName() string {
	return "matches"
}

type DB struct {
	conn *gorm.DB
}

func NewDb() *DB {
	return &DB{}
}

func (db *DB) connect(path string) error {
	conn, err := gorm.Open("sqlite3", path)
	conn.AutoMigrate(&MatchData{}, &PlayerData{})
	if err != nil {
		return err
	}
	db.conn = conn
	return nil
}

func (db *DB) newMatch() *MatchData {
	var m = MatchData{}
	db.conn.Create(&m)
	return &m
}

func (db *DB) saveOrDelMatchData(m *MatchData) {
	if m.Elasped > 0 && len(m.Member) > 0 {
		db.conn.Save(m)
	} else {
		db.conn.Delete(m)
	}
}

func (db *DB) getHistory(count int) []MatchData {
	var matches []MatchData
	db.conn.Order("id desc").Where("elasped > 0").Limit(count).Preload("Member").Find(&matches)
	return matches
}

func (db *DB) startAnswer(mid int, eid string) *MatchData {
	var match MatchData
	db.conn.Preload("Member").Where("id = ?", mid).First(&match)
	match.AnswerType = MatchAnswering
	match.ExternalID = eid
	db.conn.Save(&match)
	return &match
}

func (db *DB) stopAnswer(mid int) {
	var match MatchData
	db.conn.Model(&match).Where("id = ?", mid).Update("answer_type", MatchAnswered)
}

func (db *DB) updateQuestionInfo(pid int, qid string, aid string) *PlayerData {
	var player PlayerData
	db.conn.Where("id = ?", pid).First(&player)
	m, _ := url.ParseQuery(player.QuestionInfo)
	m.Set(qid, aid)
	player.QuestionInfo = m.Encode()
	player.Answered = len(m)
	db.conn.Save(&player)
	return &player
}

func (db *DB) updatePlayerData(pid int, name string, uid string) *PlayerData {
	var player PlayerData
	db.conn.Where("id = ?", pid).First(&player)
	player.Name = name
	player.ExternalID = uid
	db.conn.Save(&player)
	return &player
}

func (db *DB) getAnsweringMatchData() *MatchData {
	matches := db.getHistory(12)
	for _, match := range matches {
		if match.AnswerType == MatchAnswering {
			return &match
		}
	}
	return nil
}

func (db *DB) updateMatchData(mid int, eid string) *MatchData {
	var match MatchData
	db.conn.Where("id = ?", mid).First(&match)
	match.ExternalID = eid
	db.conn.Save(&match)
	return &match
}
