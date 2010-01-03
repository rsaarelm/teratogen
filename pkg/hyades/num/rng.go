package num

import (
	"exp/iterable"
	"math"
	"rand"
	"time"
)

type RandState int64

func RandomFromIterable(iter iterable.Iterable) interface{} {
	seq := iterable.Data(iter)
	return seq[rand.Intn(len(seq))]
}

func WithProb(prob float64) bool { return rand.Float64() < prob }

func OneChanceIn(num int) bool { return rand.Intn(num) == 0 }

// MakeRandState initializes the random number generator using the given value
// and returns the RandState value which can be used to return the generator
// to the same state.
func NewRandState(state int64) (result RandState) {
	result = RandState(state)
	RestoreRngState(result)
	return
}

// RngSeedFromClock seeds the random number generator from the system clock.
// Return the generator state so that the same value can be re-used if
// desired.
func RandStateFromClock() RandState { return NewRandState(time.Nanoseconds()) }

// SaveRntState generates a new random number generator state, which can be
// used to return the generator to this state. As currently implemented, this
// operation changes the state of the random number generator. It may also
// reduce entropy of the generator if used frequently.
func SaveRngState() (result RandState) {
	// Since we can't get at the actual rng state, use a trick instead.
	// Use the rng to generate a new seed value, and both seed the
	// generator to that value and return the value. Now this value can be
	// used as a checkpoint to get the rng to returning the exact same
	// subsequent results every time.
	result = RandState(rand.Int63())
	RestoreRngState(result)
	return
}

// RestoreRngState restores the state of the random number generator from the
// value stored by SaveRngState.
func RestoreRngState(state RandState) { rand.Seed(int64(state)) }

// RandomAngle returns a random angle in radians.
func RandomAngle() (radian float64) { return rand.Float64() * 2.0 * math.Pi }
