package core

import (
	"github.com/BurntSushi/toml"
	"log"
	"os"
)

var _ = log.Printf

type SurveyQuestion struct {
	Q       string   `json:"q"`
	Options []string `json:"options"`
}

type Survey struct {
	Questions []SurveyQuestion `json:"questions"`
}

var survey = LoadSurvey()

func GetSurvey() *Survey {
	return survey
}

func LoadSurvey() *Survey {
	var s Survey
	if _, err := toml.DecodeFile("survey.toml", &s); err != nil {
		log.Printf("parse survey error:%v\n", err.Error())
		os.Exit(1)
	}
	return &s
}
