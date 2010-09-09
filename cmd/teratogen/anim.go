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
	startColor image.Color
	endColor   image.Color
	size       int

	// Velocity
	dx, dy float64
	// Acceleration
	d2x, d2y float64
	lifetime int64

	life int64

	// Position
	x, y float64
}

func (self *particle) Color() image.Color {
	relativeLife := float64(self.life) / float64(self.lifetime)
	return gfx.LerpColor(self.endColor, self.startColor, relativeLife)
}

func newParticle(x, y float64, mat ParticleMaterial, emit ParticleEmitter) (result *particle) {
	track := emit.Emit()
	return &particle{
		mat.StartColor,
		mat.EndColor,
		mat.Size * VisualScale(),
		track.Dx * float64(VisualScale()), track.Dy * float64(VisualScale()),
		track.D2x * float64(VisualScale()), track.D2y * float64(VisualScale()),
		track.Lifetime,
		track.Lifetime,
		x, y}
}

type ParticleMaterial struct {
	StartColor image.Color
	EndColor   image.Color
	Size       int
}

var (
	materialSpark = ParticleMaterial{gfx.White, gfx.Yellow, 1}
	materialBlood = ParticleMaterial{gfx.Red, gfx.Red, 1}
	materialGib   = ParticleMaterial{gfx.DarkRed, gfx.DarkRed, 2}
	materialBile  = ParticleMaterial{gfx.DarkKhaki, gfx.DarkGreen, 1}
)

type ParticleTrajectory struct {
	// Velocity in units per second
	Dx, Dy float64
	// Acceleration in units per second squared
	D2x, D2y float64
	// Lifetime in nanoseconds
	Lifetime int64
}

type ParticleEmitter interface {
	Emit() ParticleTrajectory
}

type particleEmitterFn func() ParticleTrajectory

func (self particleEmitterFn) Emit() ParticleTrajectory { return self() }

func ParticleBlastEmitter(avgRadius float64, avgSpeed float64) ParticleEmitter {
	return particleEmitterFn(func() ParticleTrajectory {
		avgLife := avgRadius * 1e9 / avgSpeed

		// Perturb speed and lifetime using normal distribution.
		speed := num.Clamp(avgSpeed/4.0, avgSpeed*2.0,
			rand.NormFloat64()*math.Fabs(avgSpeed)/4.0+avgSpeed)
		lifetime := int64(rand.NormFloat64()*(avgLife/4) + avgLife)
		angle := num.RandomAngle()

		return ParticleTrajectory{
			speed * math.Cos(angle), speed * math.Sin(angle),
			0, 0,
			lifetime}
	})
}

// Blasts particles in all directions from origin.
func ParticleAnim(anim *gfx.Anim, mat ParticleMaterial, emit ParticleEmitter, x, y int, particleCount int) {
	if particleCount < 1 {
		return
	}

	defer anim.Close()
	particles := make([]*particle, particleCount)

	for i := 0; i < len(particles); i++ {
		particles[i] = newParticle(float64(x), float64(y), mat, emit)
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
				p.dx += p.d2x * float64(t) / 1e9
				p.dy += p.d2y * float64(t) / 1e9
				// XXX: Could have nicer particles.
				g.FillRect(image.Rect(int(p.x), int(p.y), int(p.x)+p.size, int(p.y)+p.size), p.Color())
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
