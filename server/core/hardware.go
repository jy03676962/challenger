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
	AnimationFakeBooks    bool
	MagicBooksLEDStatus   bool //led开关
	MagicBooksLightStatus bool //射灯开关
	Candles               map[int]int
	DoorExit              int
	LightStatus           bool
	Bgm                   int
}

//楼梯间
type Room3 struct {
	InAnimation    bool
	MagicWords     int
	Table          MagicTable
	Candles        map[int]int
	DoorExit       int
	LightStatus    bool
	LightExitStair bool
	Bgm            int
}

//魔法研究室
type Room4 struct {
	InAnimation bool
	MagicWords  int
	Table       MagicTable
	Stands      [4]MagicStands //魔法座台
	LightStatus bool
	DoorExit    int
	Bgm         int
}

//观星阁楼
type Room5 struct {
	InAnimation         bool
	MagicWords          int
	Table               MagicTable
	ConstellationSymbol map[int]int //星座符号
	LightStatus         bool
	DoorExit            int
	DoorMagicRod        int
	Bgm                 int
}

//献祭房间
type Room6 struct {
	InAnimation   bool
	MagicWords    int
	Table         MagicTable
	CurrentSymbol int
	PowerPoint    map[int]int
	LightStatus   bool
	DoorExit      int
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
	MarkAngle    int          //标记初始位置
	ButtonStatus map[int]bool "30"
	IsUseful     bool
	IsDestroyed  bool
	IsFinish     bool //可以被摧毁
}

//魔法座台
type MagicStands struct {
	IsPowerOn  bool
	IsPowerful bool
}
