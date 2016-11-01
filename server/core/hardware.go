package core

import "log"

var _ = log.Printf

type EntranceRoom struct {
	TouchMode  bool //1可用，0不可用
	DoorStatus int  //1上锁，0解锁
	LightStar  int
	Bgm        int
}

//会客室
type Room1 struct {
	MagicWords    int
	DoorMirror    int
	DoorWardrobe  int  //衣柜门
	CandleStatus  int  //蜡烛
	CrystalStatus int  //水晶
	LightStatus   bool //照明灯
	Bgm           int
}

//图书馆
type Room2 struct {
	InAnimation           bool
	MagicWords            int
	Table                 MagicTable
	FakeBooks             map[int]bool //假书no.? 是否开
	CurrentFakeBookLight  int          //已经点亮的假书
	AnimationFakeBooks    bool
	MagicBooksLEDStatus   bool //led开关
	MagicBooksLightStatus bool //射灯开关
	Candles               map[int]int
	DoorExit              int
	LightStatus           bool
	Step                  int
	Bgm                   int
}

//楼梯间
type Room3 struct {
	InAnimation         bool
	MagicWords          int
	Table               MagicTable
	Candles             map[int]int
	CurrentCandlesLight int
	DoorExit            int
	LightStatus         bool
	LightExitStair      bool
	Step                int
	Bgm                 int
}

//魔法研究室
type Room4 struct {
	InAnimation bool
	MagicWords  int
	Table       MagicTable
	Stands      [4]MagicStands //魔法座台
	LightStatus bool
	DoorExit    int
	Step        int
	Bgm         int
}

//观星阁楼
type Room5 struct {
	InAnimation               bool
	MagicWords                int
	Table                     MagicTable
	ConstellationSymbol       map[string]bool //星座符号
	CurrentConstellationLight int             //当前点亮法阵台的星座数量
	LightStatus               bool
	DoorExit                  int
	DoorMagicRod              int
	Step                      int
	Bgm                       int
}

//献祭房间
type Room6 struct {
	InAnimation   bool
	MagicWords    int
	Table         MagicTable
	NextStep      int
	CurrentSymbol int
	PowerPoint    map[int]int
	LightStatus   bool
	DoorExit      int
	Step          int
	Bgm           int
}

type ExitRoom struct {
	LightStar       int
	Bgm             int
	ButtonNextStage bool
}

//法阵
type MagicTable struct {
	CurrentAngle int
	MarkAngle    int         //标记初始位置
	ButtonStatus map[int]int "30"
	IsUseful     bool
	IsDestroyed  bool
	IsFinish     bool //可以被摧毁
}

//魔法座台
type MagicStands struct {
	IsPowerOn  bool
	IsPowerful bool
}

func NewEntranceRoom() *EntranceRoom {
	eR := EntranceRoom{}
	eR.Bgm = 0
	eR.DoorStatus = 0
	eR.LightStar = 0
	eR.TouchMode = false
	return &eR
}

func NewExitRoom() *ExitRoom {
	exR := ExitRoom{}
	exR.Bgm = 0
	exR.ButtonNextStage = false
	exR.LightStar = 0
	return &exR
}

func NewRoom1() *Room1 {
	r1 := Room1{}
	r1.Bgm = 0
	r1.CandleStatus = 0
	r1.CrystalStatus = 0
	r1.DoorMirror = 0
	r1.DoorWardrobe = 0
	r1.LightStatus = false
	return &r1
}

func NewRoom2() *Room2 {
	r2 := Room2{}
	r2.AnimationFakeBooks = false
	r2.Bgm = 0
	r2.Candles = map[int]int{
		1:  0,
		2:  0,
		3:  0,
		4:  0,
		5:  0,
		6:  0,
		7:  0,
		8:  0,
		9:  0,
		10: 0,
		11: 0,
		12: 0,
	}
	r2.DoorExit = 0
	r2.FakeBooks = map[int]bool{
		1:  false,
		2:  false,
		3:  false,
		4:  false,
		5:  false,
		6:  false,
		7:  false,
		8:  false,
		9:  false,
		10: false,
		11: false,
		12: false,
		13: false,
		14: false,
		15: false,
	}
	r2.CurrentFakeBookLight = 0
	r2.InAnimation = false
	r2.LightStatus = false
	r2.MagicBooksLEDStatus = false
	r2.MagicBooksLightStatus = false
	r2.MagicWords = 0
	r2.Table = MagicTable{}
	r2.Table.ButtonStatus = map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
	}
	r2.Table.CurrentAngle = 0
	r2.Table.IsFinish = false
	r2.Table.IsUseful = false
	r2.Table.IsDestroyed = false
	r2.Step = 1
	return &r2
}

func NewRoom3() *Room3 {
	r3 := Room3{}
	r3.Bgm = 0
	r3.Candles = map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
	}
	r3.CurrentCandlesLight = 6
	r3.DoorExit = 0
	r3.InAnimation = false
	r3.LightStatus = false
	r3.LightExitStair = false
	r3.MagicWords = 0
	r3.Table = MagicTable{}
	r3.Table.IsFinish = false
	r3.Table.IsUseful = false
	r3.Table.IsDestroyed = false
	r3.Step = 1
	return &r3
}

func NewRoom4() *Room4 {
	r4 := Room4{}
	r4.Bgm = 0
	r4.DoorExit = 0
	r4.InAnimation = false
	r4.LightStatus = false
	r4.MagicWords = 0
	for i := 0; i < 4; i++ {
		r4.Stands[i].IsPowerful = false
		r4.Stands[i].IsPowerOn = false
	}
	r4.Table = MagicTable{}
	r4.Table.IsFinish = false
	r4.Table.IsUseful = false
	r4.Table.IsDestroyed = false
	return &r4
}

func NewRoom5() *Room5 {
	r5 := Room5{}
	r5.Bgm = 0
	r5.CurrentConstellationLight = 0
	r5.ConstellationSymbol = map[string]bool{
		"A": false,
		"B": false,
		"C": false,
		"D": false,
	}
	r5.DoorExit = 0
	r5.DoorMagicRod = 0
	r5.InAnimation = false
	r5.LightStatus = false
	r5.MagicWords = 0
	r5.Table = MagicTable{}
	r5.Table.ButtonStatus = map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
	}
	r5.Table.IsFinish = false
	r5.Table.IsUseful = false
	r5.Table.IsDestroyed = false
	return &r5
}

func NewRoom6() *Room6 {
	r6 := Room6{}
	r6.Bgm = 0
	r6.CurrentSymbol = 0
	r6.DoorExit = 0
	r6.InAnimation = false
	r6.LightStatus = false
	r6.MagicWords = 0
	r6.NextStep = 0
	r6.PowerPoint = map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
	}
	r6.Table = MagicTable{}
	r6.Table.ButtonStatus = map[int]int{
		1: 0,
		2: 0,
		3: 0,
		4: 0,
		5: 0,
		6: 0,
	}
	return &r6
}