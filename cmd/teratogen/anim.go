package main

import (
	"exp/draw"
	"hyades/gfx"
	"hyades/num"
	"image"
	"math"
	"rand"
)

func TestAnim1(anim *gfx.Anim) {
	defer anim.Close()
	t := int64(0)
	for t < 2e9 {
		g, dt := anim.StartDraw()
		t += dt
		for x := 0; x < g.Width(); x++ {
			h := float64(g.Height())
			w := float64(g.Width())
			y := int(h/2 + h/4*math.Sin(float64(t)/1e8+float64(x)/w*16))
			g.Set(x, y, gfx.AliceBlue)
		}
		anim.StopDraw()
	}
}

func TestAnim2(anim *gfx.Anim) {
	defer anim.Close()
	t := int64(0)
	for t < 2e9 {
		g, dt := anim.StartDraw()
		t += dt
		gfx.ThickLine(g, draw.Pt(0, 0), draw.Pt(1000, 1000), gfx.Teal, config.Scale*2)
		anim.StopDraw()
	}
}

type particle struct {
	x, y, dx, dy float64
	startColor   image.Color
	endColor     image.Color
	life         int64
	lifetime     int64
}

func newParticle(x, y int, lifetime int64, speed float64, startColor, endColor image.Color) (result *particle) {
	result = new(particle)

	result.x, result.y = float64(x), float64(y)

	// Perturb speed and lifetime using normal distribution.
	speed = num.Clamp(speed/4.0, speed*2.0, rand.NormFloat64()*math.Fabs(speed)/4.0+speed)
	result.lifetime = int64(rand.NormFloat64()*float64(lifetime/4) + float64(lifetime))
	result.life = result.lifetime
	result.startColor, result.endColor = startColor, endColor
	angle := num.RandomAngle()
	result.dx = speed * math.Cos(angle)
	result.dy = speed * math.Sin(angle)

	return
}

func (self *particle) Color() image.Color {
	relativeLife := float64(self.life) / float64(self.lifetime)
	return gfx.LerpColor(self.endColor, self.startColor, relativeLife)
}

// Blasts particles in all directions from origin.
func ParticleAnim(anim *gfx.Anim, x, y int, size int, lifetime int64, speed float64, startColor, endColor image.Color, particleCount int) {
	defer anim.Close()
	particles := make([]*particle, particleCount)

	for i := 0; i < len(particles); i++ {
		particles[i] = newParticle(x, y, lifetime, speed, startColor, endColor)
	}

	liveOnes := len(particles)
	for liveOnes > 0 {
		g, t := anim.StartDraw()

		liveOnes = 0
		for _, p := range particles {
			p.life = p.life - t
			if p.life > 0 {
				liveOnes++
				p.x += p.dx * float64(t) / 1e9
				p.y += p.dy * float64(t) / 1e9
				// XXX: Could have nicer particles.
				g.FillRect(draw.Rect(int(p.x), int(p.y), int(p.x)+size, int(p.y)+size), p.Color())
			}
		}
		anim.StopDraw()
	}
}

func LineAnim(anim *gfx.Anim, p1, p2 draw.Point, lifetime int64, startColor, endColor image.Color, thickness int) {
	defer anim.Close()
	life := lifetime
	for life > 0 {
		g, t := anim.StartDraw()
		life -= t
		col := gfx.LerpColor(endColor, startColor, float64(life)/float64(lifetime))
		if thickness == 1 {
			gfx.Line(g, p1, p2, col)
		} else {
			gfx.ThickLine(g, p1, p2, col, thickness)
		}
		anim.StopDraw()
	}
}
