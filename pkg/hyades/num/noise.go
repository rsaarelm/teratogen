package num

import (
	"math"
)

// Noise generates pseudorandom noise values. From Hugo Elias,
// http://freespace.virgin.net/hugo.elias/models/m_perlin.htm
func Noise(seed int) float64 {
	seed = (seed << 13) ^ seed
	return (1.0 -
		float64((seed*(seed*seed*15731+789221)+1376312589)&0x7fffffff)/
			1073741824.0)
}

// SmoothNoise3D generates pseudorandom smoothed noise in 3D space. From Ken
// Perlin's page, http://mrl.nyu.edu/~perlin/noise/
func SmoothNoise3D(x, y, z float64) float64 {
	// Get unit cube position in byte-3 space.
	cubeX := int(math.Floor(x)) & 0xff
	cubeY := int(math.Floor(y)) & 0xff
	cubeZ := int(math.Floor(z)) & 0xff

	// Get the position within the unit cube.
	x, y, z = math.Fmod(x, 1.0), math.Fmod(y, 1.0), math.Fmod(z, 1.0)

	// Fade curves for x, y, z
	u, v, w := fadeCurve(x), fadeCurve(y), fadeCurve(z)

	// Cube corner hash coordniates
	a := noiseSeed[cubeX] + cubeY
	aa, ab := noiseSeed[a]+cubeZ, noiseSeed[a+1]+cubeZ
	b := noiseSeed[cubeX+1] + cubeY
	ba, bb := noiseSeed[b]+cubeZ, noiseSeed[b+1]+cubeZ

	// Interpolate cube corners
	edge1 := Lerp(perlinGrad(noiseSeed[aa], x, y, z),
		perlinGrad(noiseSeed[ba], x-1, y, z), u)
	edge2 := Lerp(perlinGrad(noiseSeed[ab], x, y-1, z),
		perlinGrad(noiseSeed[bb], x-1, y-1, z), u)
	edge3 := Lerp(perlinGrad(noiseSeed[aa], x, y, z-1),
		perlinGrad(noiseSeed[ba], x-1, y, z-1), u)
	edge4 := Lerp(perlinGrad(noiseSeed[ab], x, y-1, z-1),
		perlinGrad(noiseSeed[bb], x-1, y-1, z-1), u)

	return Lerp(Lerp(edge1, edge2, v), Lerp(edge3, edge4, v), w)
}

// PerlinNoise generates Perlin noise, pseudorandom interpolated noise in a 3D
// space. This is based on Ken Perlin's improved 2002 implementation of the
// algorithm.
func PerlinNoise(persistence float64, octaves int, x, y, z float64) (result float64) {
	for i := 0; i < octaves; i++ {
		freq := math.Pow(2, float64(i))
		amp := math.Pow(persistence, float64(i))
		result += SmoothNoise3D(x*freq, y*freq, z*freq) * amp
	}
	return
}

func fadeCurve(t float64) float64 { return t * t * t * (t*(t*6-15) + 10) }

// perlinGrad returns the gradient of a vector relative to an edge of the unit
// cube psedorandomly chosen based on the hash. Used by Perlin's smooth noise
// function.
func perlinGrad(hash int, x, y, z float64) (result float64) {
	// Convert the low 4 bits of the hash code into one of 12 gradient directions
	h := hash & 0xf
	var u, v float64
	if h < 8 {
		u = x
	} else {
		u = y
	}
	if h < 4 {
		v = y
	} else {
		if h == 12 || h == 14 {
			v = x
		} else {
			v = z
		}
	}
	if h&1 == 0 {
		result = u
	} else {
		result = -u
	}
	if h&2 == 0 {
		result += v
	} else {
		result -= v
	}
	return
}

// Init seed table for Perlin's noise function. Values from the source
// code at Ken Perlin's page.
var noiseSeed = [512]int{151, 160, 137, 91, 90, 15,
	131, 13, 201, 95, 96, 53, 194, 233, 7, 225, 140, 36, 103, 30, 69, 142, 8, 99, 37, 240, 21, 10, 23,
	190, 6, 148, 247, 120, 234, 75, 0, 26, 197, 62, 94, 252, 219, 203, 117, 35, 11, 32, 57, 177, 33,
	88, 237, 149, 56, 87, 174, 20, 125, 136, 171, 168, 68, 175, 74, 165, 71, 134, 139, 48, 27, 166,
	77, 146, 158, 231, 83, 111, 229, 122, 60, 211, 133, 230, 220, 105, 92, 41, 55, 46, 245, 40, 244,
	102, 143, 54, 65, 25, 63, 161, 1, 216, 80, 73, 209, 76, 132, 187, 208, 89, 18, 169, 200, 196,
	135, 130, 116, 188, 159, 86, 164, 100, 109, 198, 173, 186, 3, 64, 52, 217, 226, 250, 124, 123,
	5, 202, 38, 147, 118, 126, 255, 82, 85, 212, 207, 206, 59, 227, 47, 16, 58, 17, 182, 189, 28, 42,
	223, 183, 170, 213, 119, 248, 152, 2, 44, 154, 163, 70, 221, 153, 101, 155, 167, 43, 172, 9,
	129, 22, 39, 253, 19, 98, 108, 110, 79, 113, 224, 232, 178, 185, 112, 104, 218, 246, 97, 228,
	251, 34, 242, 193, 238, 210, 144, 12, 191, 179, 162, 241, 81, 51, 145, 235, 249, 14, 239, 107,
	49, 192, 214, 31, 181, 199, 106, 157, 184, 84, 204, 176, 115, 121, 50, 45, 127, 4, 150, 254,
	138, 236, 205, 93, 222, 114, 67, 29, 24, 72, 243, 141, 128, 195, 78, 66, 215, 61, 156, 180}

func init() {
	for i := 0; i < 256; i++ {
		noiseSeed[256+i] = noiseSeed[i]
	}
}
