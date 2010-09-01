package main

import (
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
		for x := g.Bounds().Min.X; x < g.Bounds().Max.X; x++ {
			h := float64(g.Bounds().Dy())
			w := float64(g.Bounds().Dx())
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
		gfx.ThickLine(g, image.Pt(0, 0), image.Pt(1000, 1000), gfx.Teal, 2)
		anim.StopDraw()
	}
}

type particle struct {
	// Position
	x, y float64
	// Velocity
	dx, dy     float64
	startColor image.Color
	endColor   image.Color
	life       int64
	lifetime   int64
	size       int
}

type ParticleMaterial struct {
	StartColor image.Color
	EndColor   image.Color
	Size       int
}

var (
	materialSpark = ParticleMaterial{gfx.White, gfx.Yellow, VisualScale()}
	materialBlood = ParticleMaterial{gfx.Red, gfx.Red, VisualScale()}
	materialGib   = ParticleMaterial{gfx.Red, gfx.Red, VisualScale() * 2}
	materialBile  = ParticleMaterial{gfx.DarkKhaki, gfx.DarkGreen, VisualScale()}
)

type ParticleTrajectory struct {
	Dx, Dy   float64
	Lifetime int64
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
func ParticleAnim(anim *gfx.Anim, mat ParticleMaterial, x, y int, lifetime int64, speed float64, particleCount int) {
	defer anim.Close()
	particles := make([]*particle, particleCount)

	for i := 0; i < len(particles); i++ {
		particles[i] = newParticle(x, y, lifetime, speed, mat.StartColor, mat.EndColor)
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
				g.FillRect(image.Rect(int(p.x), int(p.y), int(p.x)+mat.Size, int(p.y)+mat.Size), p.Color())
			}
		}
		anim.StopDraw()
	}
}

func LineAnim(anim *gfx.Anim, p1, p2 image.Point, lifetime int64, startColor, endColor image.Color, thickness int) {
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
