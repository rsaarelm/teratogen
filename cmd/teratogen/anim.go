package main

import (
	"hyades/gfx"
	"hyades/sdl"
	"math"
)

type Anim struct {
	// Channel that receives an interval in nanoseconds since the last
	// update, causes the anim to draw itself.
	UpdateChan chan int64
	// The position of the animation in the draw queue. Low z values are
	// drawn first.
	Z float64
}

func NewAnim(z float64) *Anim { return &Anim{make(chan int64), z} }

func (self *Anim) Close() {
	close(self.UpdateChan)
	<-self.UpdateChan
}

func (self *Anim) Closed() bool { return closed(self.UpdateChan) }

func TestAnim(context sdl.Context, anim *Anim) {
	defer anim.Close()
	t := int64(0)
	for t < 2e9 {
		t += <-anim.UpdateChan
		scr := context.Screen()
		col, _ := gfx.ParseColor("AliceBlue")
		for x := 0; x < scr.Width(); x++ {
			h := float64(scr.Height())
			w := float64(scr.Width())
			y := int(h/2 + h/4*math.Sin(float64(t)/1e8+float64(x)/w*16))
			scr.Set(x, y, col)
		}
		anim.UpdateChan <- 0
	}
}
