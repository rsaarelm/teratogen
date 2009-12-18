package sfx

import (
	. "hyades/common"
	"hyades/num"
	"math"
)

type freqWaveFunc func(t float, hz float) float

// 1 Hz sine wave
func Sine(t float, hz float) float {
	return float(math.Sin(float64(t * hz * 2.0 * math.Pi)))
}

// Make a pulse wave that spends duty = (0..1) in active phase.
func MakePulse(duty float) freqWaveFunc {
	Assert(duty > 0.0 && duty < 1.0, "Invalid pulse wave duty %v.", duty)
	return func(t float, hz float) float {
		t = float(num.Fracf(float64(t * hz)))
		if t < duty {
			return -1.0
		}
		return 1.0
	}
}

// Square wave, wave = -1.0 from t 0.0..0.5, wave = 1.0 from t 0.5..1.0, then
// repeat.
func Square(t float, hz float) float {
	t = float(num.Fracf(float64(t * hz)))
	if t < 0.5 {
		return -1.0
	}
	return 1.0
}

func Triangle(t float, hz float) float {
	t = float(num.Fracf(float64(t * hz)))
	if t < 0.5 {
		return -1.0 + t * 4
	}
	t -= 0.5
	return 1.0 - (t * 4)
}

func Sawtooth(t float, hz float) float {
	t = float(num.Fracf(float64(t * hz)))
	return t
}

// Noise wave is like a square wave with random levels instead of 1.0, -1.0.
// Change frequency to modify the sound.
func Noise(t float, hz float) float {
	return float(num.Noise(int(t * hz * 2.0)))
}

// Offsets the wave frequency by deltaHz from time t onward
func Jump(jumpT float, deltaHz float, wave freqWaveFunc) freqWaveFunc {
	return func(t float, hz float) float {
		if t > jumpT { hz += deltaHz }
		return wave(t, hz)
	}
}

// Slides the frequency by velocity along time. Alters velocity by
// acceleration by time.
func Slide(velocity float, acceleration float, wave freqWaveFunc) freqWaveFunc {
	return func(t float, hz float) float {
		hz += t * velocity + 0.5 * t * t * acceleration
		return wave(t, hz)
	}
}

func MakeWave(hz float, fn freqWaveFunc) WaveFunc {
	return func (t float) float {
		return fn(t, hz)
	}
}

func AmpFilter(amplitude float, wave WaveFunc) WaveFunc {
	return func(t float) float {
		return amplitude * wave(t)
	}
}

// Attack-decay-sustain-release wave envelope
func ADSRFilter(attackTime float, decayTime float, sustainLevel float,
	sustainTime float, releaseTime float, wave WaveFunc) WaveFunc {
	return func(t float) float {
		var amp float
		if attackTime > 0.0 && t < attackTime {
			amp = t / attackTime
		} else {
			t -= attackTime
			if decayTime > 0.0 && t < decayTime {
				amp = 1.0 - t / decayTime * (1.0 - sustainLevel)
			} else {
				t -= decayTime
				if t < sustainTime {
					amp = sustainLevel
				} else {
					t -= sustainTime
					if releaseTime > 0.0 && t < releaseTime {
						amp = sustainLevel - t / releaseTime * sustainLevel
					} else { amp = 0.0 } } } }
		return amp * wave(t)
	}
}
