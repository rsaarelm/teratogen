package gamelib

import "exp/iterable"
import "rand"

func RandomFromIterable(iter iterable.Iterable) interface{} {
	seq := iterable.Data(iter);
	return seq[rand.Intn(len(seq))];
}

func WithProb(prob float64) bool {
	return rand.Float64() < prob;
}

func OneChanceIn(num int) bool {
	return rand.Intn(num) == 0;
}