package core

type LaserInterface interface {
	Pause(t float64) int
	IsFollow(cid string) bool
	Tick(dt float64)
	Close()
}
