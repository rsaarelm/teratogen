package main

import (
	"exp/draw"
	"hyades/gfx"
	"hyades/num"
	"hyades/sdl"
	"image"
	"math"
	"rand"
)

func TestAnim(context sdl.Context, anim *gfx.Anim) {
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
func ParticleAnim(context sdl.Context, anim *gfx.Anim, x, y int, size int, lifetime int64, speed float64, color image.Color, particleCount int) {
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
				g.FillRect(draw.Rect(int(p.x), int(p.y), int(p.x)+size, int(p.y)+size), p.color)
			}
		}
		anim.StopDraw()
	}
}
