package main

import (
	"exp/draw"
	"hyades/gfx"
	"hyades/gui"
	"hyades/num"
	"hyades/sdl"
	"image"
	"math"
	"rand"
)

// State wrapper structure for animation objects. Animations are not a proper
// part of the game world, they don't go into savegames, but they can be used
// to illustrate events in the world. Animations are run in separate
// goroutines, which communicate with the main game via the Anim object API.
type Anim struct {
	// Channel that receives an interval in nanoseconds since the last
	// update, causes the anim to draw itself.
	updateChan chan int64

	// The current graphics context, passed using the Update function.
	g gui.Graphics

	// The position of the animation in the draw queue. Low z values are
	// drawn first.
	Z float64
}

func NewAnim(z float64) *Anim { return &Anim{make(chan int64), nil, z} }

// Update causes the anim to draw itself in the graphics context and advance
// it's state by elapsedNs. Calling with 0 elapsedNs will just make the anim
// redraw itself.
func (self *Anim) Update(g gui.Graphics, elapsedNs int64) {
	self.g = g
	self.updateChan <- elapsedNs
	// Wait for the draw to finish
	<-self.updateChan
	// Just to make sure it isn't used inappropriately.
	self.g = nil
}

func (self *Anim) Close() {
	close(self.updateChan)
	<-self.updateChan
}

func (self *Anim) Closed() bool { return closed(self.updateChan) }

// StartDraw is called from the animation goroutine. It waits until the main
// engine wants the animation to draw itself. It returns the graphics context
// to draw into and the number of nanoseconds that the animation state should
// advance. Do not call this outside the animation goroutine.
func (self *Anim) StartDraw() (g gui.Graphics, elapsedNs int64) {
	elapsedNs = <-self.updateChan
	g = self.g
	return
}

// StopDraw is called in from the animation goroutine when the animation has
// finished drawing its current state after calling StartDraw.
func (self *Anim) StopDraw() { self.updateChan <- 0 }

func TestAnim(context sdl.Context, anim *Anim) {
	defer anim.Close()
	t := int64(0)
	for t < 2e9 {
		g, t := anim.StartDraw()
		col, _ := gfx.ParseColor("AliceBlue")
		for x := 0; x < g.Width(); x++ {
			h := float64(g.Height())
			w := float64(g.Width())
			y := int(h/2 + h/4*math.Sin(float64(t)/1e8+float64(x)/w*16))
			g.Set(x, y, col)
		}
		anim.StopDraw()
	}
}

type particle struct {
	x, y, dx, dy float64
	color        image.Color
	life         int64
}

func newParticle(x, y int, lifetime int64, speed float64, color image.Color) (result *particle) {
	result = new(particle)

	result.x, result.y = float64(x), float64(y)

	// Perturb speed and lifetime using normal distribution.
	speed = num.Clamp(speed/4.0, speed*2.0, rand.NormFloat64()*math.Fabs(speed)/4.0+speed)
	result.life = int64(rand.NormFloat64()*float64(lifetime/4) + float64(lifetime))
	result.color = color
	angle := num.RandomAngle()
	result.dx = speed * math.Cos(angle)
	result.dy = speed * math.Sin(angle)

	return
}

// Blasts particles in all directions from origin.
func ParticleAnim(context sdl.Context, anim *Anim, x, y int, lifetime int64, speed float64, color image.Color, particleCount int) {
	defer anim.Close()
	particles := make([]*particle, particleCount)

	for i := 0; i < len(particles); i++ {
		particles[i] = newParticle(x, y, lifetime, speed, color)
	}

	liveOnes := len(particles)
	for liveOnes > 0 {
		g, t := anim.StartDraw()

		liveOnes = 0
		for _, p := range particles {
			if p.life > 0 {
				p.life = p.life - t
				liveOnes++
				p.x += p.dx * float64(t) / 1e9
				p.y += p.dy * float64(t) / 1e9
				// XXX: Could have nicer particles.
				g.FillRect(draw.Rect(int(p.x), int(p.y), int(p.x)+2, int(p.y)+2), p.color)
			}
		}
		anim.StopDraw()
	}
}
