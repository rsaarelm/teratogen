package num

import ()

type weightNode struct {
	weight float64
	value  interface{}
}

type WeightedDist []weightNode

func MakeWeightedDist(weight func(val interface{}) float64, values []interface{}) (result WeightedDist) {
	result = make([]weightNode, len(values))
	sum := result.gather(weight, values)
	result.normalize(sum)
	return
}

func (self WeightedDist) gather(weight func(val interface{}) float64, values []interface{}) (sum float64) {
	for i := 0; i < len(values); i++ {
		w := weight(values[i])
		sum += w
		self[i] = weightNode{w, values[i]}
	}
	return
}

func (self WeightedDist) normalize(sum float64) {
	for i := 0; i < len(self); i++ {
		self[i].weight /= sum
	}
}

// Sample takes a value in [0, 1) and returns the item from the distribution
// on whose slice the value falls. The lengths of item slices are proportional
// to their weights.
func (self WeightedDist) Sample(x float64) interface{} {
	var sum float64 = 0
	for _, item := range self {
		sum += item.weight
		if sum >= x {
			return item.value
		}
	}
	return self[len(self)-1].value
}

func (self WeightedDist) Empty() bool { return len(self) == 0 }
