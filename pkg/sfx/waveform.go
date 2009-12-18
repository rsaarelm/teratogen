package sfx

import (
	"hyades/dbg"
	"hyades/num"
	"math"
)

type freqWaveFunc func(t float64, hz float64) float64

// 1 Hz sine wave
func Sine(t float64, hz float64) float64 {
	return math.Sin(t * hz * 2.0 * math.Pi)
}

// Make a pulse wave that spends duty = (0..1) in active phase.
func MakePulse(duty float64) freqWaveFunc {
	dbg.Assert(duty > 0.0 && duty < 1.0, "Invalid pulse wave duty %v.", duty)
	return func(t float64, hz float64) float64 {
		t = float64(num.Fracf(float64(t * hz)))
		if t < duty {
			return -1.0
		}
		return 1.0
	}
}

// Square wave, wave = -1.0 from t 0.0..0.5, wave = 1.0 from t 0.5..1.0, then
// repeat.
func Square(t float64, hz float64) float64 {
	t = num.Fracf(t * hz)
	if t < 0.5 {
		return -1.0
	}
	return 1.0
}

func Triangle(t float64, hz float64) float64 {
	t = num.Fracf(t * hz)
	if t < 0.5 {
		return -1.0 + t * 4
	}
	t -= 0.5
	return 1.0 - (t * 4)
}

func Sawtooth(t float64, hz float64) float64 {
	t = num.Fracf(t * hz)
	return t
}

// Noise wave is like a square wave with random levels instead of 1.0, -1.0.
// Change frequency to modify the sound.
func Noise(t float64, hz float64) float64 {
	return num.Noise(int(t * hz * 2.0))
}

// Offsets the wave frequency by deltaHz from time t onward
func Jump(jumpT float64, deltaHz float64, wave freqWaveFunc) freqWaveFunc {
	return func(t float64, hz float64) float64 {
		if t > jumpT { hz += deltaHz }
		return wave(t, hz)
	}
}

// Slides the frequency by velocity along time. Alters velocity by
// acceleration by time. Frequency won't go below minHz.
func Slide(velocity float64, acceleration float64, minHz float64, wave freqWaveFunc) freqWaveFunc {
	return func(t float64, hz float64) float64 {
		hz += t * velocity + 0.5 * t * t * acceleration
		if hz < minHz { hz = minHz }
		return wave(t, hz)
	}
}

func MakeWave(hz float64, fn freqWaveFunc) WaveFunc {
	return func (t float64) float64 {
		return fn(t, hz)
	}
}

func AmpFilter(amplitude float64, wave WaveFunc) WaveFunc {
	return func(t float64) float64 {
		return amplitude * wave(t)
	}
}

// Attack-decay-sustain-release wave envelope
func ADSRFilter(attackTime float64, decayTime float64, sustainLevel float64,
	sustainTime float64, releaseTime float64, wave WaveFunc) WaveFunc {
	return func(t float64) float64 {
		var amp float64
		t2 := t
		if attackTime > 0.0 && t2 < attackTime {
			amp = t2 / attackTime
		} else {
			t2 -= attackTime
			if decayTime > 0.0 && t2 < decayTime {
				amp = 1.0 - t2 / decayTime * (1.0 - sustainLevel)
			} else {
				t2 -= decayTime
				if t2 < sustainTime {
					amp = sustainLevel
				} else {
					t2 -= sustainTime
					if releaseTime > 0.0 && t2 < releaseTime {
						amp = sustainLevel - t2 / releaseTime * sustainLevel
					} else { amp = 0.0 } } } }
		return amp * wave(t)
	}
}
