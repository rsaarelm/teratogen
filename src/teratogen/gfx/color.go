/* color.go

   Copyright (C) 2012 Risto Saarelma

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU General Public License as published by
   the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU General Public License for more details.

   You should have received a copy of the GNU General Public License
   along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package gfx

import (
	"image/color"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// RGBA8Bit returns the components of a color value compressed to 8 bits.
func RGBA8Bit(col color.Color) (r8, g8, b8, a8 uint8) {
	r, g, b, a := col.RGBA()
	return uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)
}

// LerpCol returns a linearly interpolated color between the two endpoint
// colors.
func LerpCol(c1, c2 color.Color, x float64) color.Color {
	r1, b1, g1, a1 := c1.RGBA()
	r2, b2, g2, a2 := c2.RGBA()

	return color.RGBA{
		lerpComponent(r1, r2, x),
		lerpComponent(g1, g2, x),
		lerpComponent(b1, b2, x),
		lerpComponent(a1, a2, x)}
}

func ScaleCol(col color.Color, scale float64) (result color.Color) {
	r, g, b, a := col.RGBA()
	r8 := uint8(math.Min(float64(r)*scale, 0xffff) / 256)
	g8 := uint8(math.Min(float64(g)*scale, 0xffff) / 256)
	b8 := uint8(math.Min(float64(b)*scale, 0xffff) / 256)
	return color.RGBA{r8, g8, b8, uint8(a >> 8)}
}

func lerpComponent(a, b uint32, x float64) uint8 {
	return uint8((float64(a) + (float64(b)-float64(a))*x) / 256)
}

var longHexCol = regexp.MustCompile("#([0-9a-fA-F]{2})([0-9a-fA-F]{2})([0-9a-fA-F]{2})")
var shortHexCol = regexp.MustCompile("#([0-9a-fA-F])([0-9a-fA-F])([0-9a-fA-F])")

// ParseColor returns a color parsed from the given string. The string can be
// of the form "#rrggbb", where rr, gg, and bb are hex color codes, "#rgb",
// where r, g, and b are hex color codes that are expanded by doubling the
// digit (0xa becomes 0xaa), or a named color like "red" or "aliceblue"
// (matching is case-insensitive) from the list of known named colors.
func ParseColor(str string) (result color.Color, ok bool) {
	str = strings.ToLower(str)
	if result, ok = namedColors[str]; ok {
		return
	}

	match := longHexCol.FindSubmatch([]byte(str))
	if match == nil {
		match = shortHexCol.FindSubmatch([]byte(str))
	}
	if match != nil && string(match[0]) == str {
		rgb := []int64{0, 0, 0}
		for i := 0; i < 3; i++ {
			hex := ""
			for len(hex) < 2 {
				hex = hex + string(match[i+1])
			}
			rgb[i], _ = strconv.ParseInt(hex, 16, 32)
		}
		return color.RGBA{uint8(rgb[0]), uint8(rgb[1]), uint8(rgb[2]), 0xff}, true
	}

	return
}

var namedColors = map[string]color.Color{
	"aliceblue":            AliceBlue,
	"antiquewhite":         AntiqueWhite,
	"aqua":                 Aqua,
	"aquamarine":           Aquamarine,
	"azure":                Azure,
	"beige":                Beige,
	"bisque":               Bisque,
	"black":                Black,
	"blanchedalmond":       BlanchedAlmond,
	"blue":                 Blue,
	"blueviolet":           BlueViolet,
	"brown":                Brown,
	"burlywood":            BurlyWood,
	"cadetblue":            CadetBlue,
	"chartreuse":           Chartreuse,
	"chocolate":            Chocolate,
	"coral":                Coral,
	"cornflowerblue":       CornflowerBlue,
	"cornsilk":             Cornsilk,
	"crimson":              Crimson,
	"cyan":                 Cyan,
	"darkblue":             DarkBlue,
	"darkcyan":             DarkCyan,
	"darkgoldenrod":        DarkGoldenRod,
	"darkgray":             DarkGray,
	"darkgreen":            DarkGreen,
	"darkkhaki":            DarkKhaki,
	"darkmagenta":          DarkMagenta,
	"darkolivegreen":       DarkOliveGreen,
	"darkorange":           DarkOrange,
	"darkorchid":           DarkOrchid,
	"darkred":              DarkRed,
	"darksalmon":           DarkSalmon,
	"darkseagreen":         DarkSeaGreen,
	"darkslateblue":        DarkSlateBlue,
	"darkslategray":        DarkSlateGray,
	"darkturquoise":        DarkTurquoise,
	"darkviolet":           DarkViolet,
	"deeppink":             DeepPink,
	"deepskyblue":          DeepSkyBlue,
	"dimgray":              DimGray,
	"dodgerblue":           DodgerBlue,
	"firebrick":            FireBrick,
	"floralwhite":          FloralWhite,
	"forestgreen":          ForestGreen,
	"fuchsia":              Fuchsia,
	"gainsboro":            Gainsboro,
	"ghostwhite":           GhostWhite,
	"gold":                 Gold,
	"goldenrod":            Goldenrod,
	"gray":                 Gray,
	"green":                Green,
	"greenyellow":          GreenYellow,
	"honeydew":             Honeydew,
	"hotpink":              HotPink,
	"indianred":            IndianRed,
	"indigo":               Indigo,
	"ivory":                Ivory,
	"khaki":                Khaki,
	"lavender":             Lavender,
	"lavenderblush":        LavenderBlush,
	"lawngreen":            LawnGreen,
	"lemonchiffon":         LemonChiffon,
	"lightblue":            LightBlue,
	"lightcoral":           LightCoral,
	"lightcyan":            LightCyan,
	"lightgoldenrodyellow": LightGoldenrodYellow,
	"lightgreen":           LightGreen,
	"lightgrey":            LightGrey,
	"lightpink":            LightPink,
	"lightsalmon":          LightSalmon,
	"lightseagreen":        LightSeaGreen,
	"lightskyblue":         LightSkyBlue,
	"lightslategray":       LightSlateGray,
	"lightsteelblue":       LightSteelBlue,
	"lightyellow":          LightYellow,
	"lime":                 Lime,
	"limegreen":            LimeGreen,
	"linen":                Linen,
	"magenta":              Magenta,
	"maroon":               Maroon,
	"mediumaquamarine":     MediumAquamarine,
	"mediumblue":           MediumBlue,
	"mediumorchid":         MediumOrchid,
	"mediumpurple":         MediumPurple,
	"mediumseagreen":       MediumSeaGreen,
	"mediumslateblue":      MediumSlateBlue,
	"mediumspringgreen":    MediumSpringGreen,
	"mediumturquoise":      MediumTurquoise,
	"mediumvioletred":      MediumVioletRed,
	"midnightblue":         MidnightBlue,
	"mintcream":            MintCream,
	"mistyrose":            MistyRose,
	"moccasin":             Moccasin,
	"navajowhite":          NavajoWhite,
	"navy":                 Navy,
	"oldlace":              OldLace,
	"olive":                Olive,
	"olivedrab":            OliveDrab,
	"orange":               Orange,
	"orangered":            OrangeRed,
	"orchid":               Orchid,
	"palegoldenrod":        PaleGoldenrod,
	"palegreen":            PaleGreen,
	"paleturquoise":        PaleTurquoise,
	"palevioletred":        PaleVioletRed,
	"papayawhip":           PapayaWhip,
	"peachpuff":            PeachPuff,
	"peru":                 Peru,
	"pink":                 Pink,
	"plum":                 Plum,
	"powderblue":           PowderBlue,
	"purple":               Purple,
	"red":                  Red,
	"rosybrown":            RosyBrown,
	"royalblue":            RoyalBlue,
	"saddlebrown":          SaddleBrown,
	"salmon":               Salmon,
	"sandybrown":           SandyBrown,
	"seagreen":             SeaGreen,
	"seashell":             Seashell,
	"sienna":               Sienna,
	"silver":               Silver,
	"skyblue":              SkyBlue,
	"slateblue":            SlateBlue,
	"slategray":            SlateGray,
	"snow":                 Snow,
	"springgreen":          SpringGreen,
	"steelblue":            SteelBlue,
	"tan":                  Tan,
	"teal":                 Teal,
	"thistle":              Thistle,
	"tomato":               Tomato,
	"turquoise":            Turquoise,
	"violet":               Violet,
	"wheat":                Wheat,
	"white":                White,
	"whitesmoke":           WhiteSmoke,
	"yellow":               Yellow,
	"yellowgreen":          YellowGreen,
}

var (
	AliceBlue            = color.RGBA{0xF0, 0xF8, 0xFF, 0xFF}
	AntiqueWhite         = color.RGBA{0xFA, 0xEB, 0xD7, 0xFF}
	Aqua                 = color.RGBA{0x00, 0xFF, 0xFF, 0xFF}
	Aquamarine           = color.RGBA{0x7F, 0xFF, 0xD4, 0xFF}
	Azure                = color.RGBA{0xF0, 0xFF, 0xFF, 0xFF}
	Beige                = color.RGBA{0xF5, 0xF5, 0xDC, 0xFF}
	Bisque               = color.RGBA{0xFF, 0xE4, 0xC4, 0xFF}
	Black                = color.RGBA{0x00, 0x00, 0x00, 0xFF}
	BlanchedAlmond       = color.RGBA{0xFF, 0xEB, 0xCD, 0xFF}
	Blue                 = color.RGBA{0x00, 0x00, 0xFF, 0xFF}
	BlueViolet           = color.RGBA{0x8A, 0x2B, 0xE2, 0xFF}
	Brown                = color.RGBA{0xA5, 0x2A, 0x2A, 0xFF}
	BurlyWood            = color.RGBA{0xDE, 0xB8, 0x87, 0xFF}
	CadetBlue            = color.RGBA{0x5F, 0x9E, 0xA0, 0xFF}
	Chartreuse           = color.RGBA{0x7F, 0xFF, 0x00, 0xFF}
	Chocolate            = color.RGBA{0xD2, 0x69, 0x1E, 0xFF}
	Coral                = color.RGBA{0xFF, 0x7F, 0x50, 0xFF}
	CornflowerBlue       = color.RGBA{0x64, 0x95, 0xED, 0xFF}
	Cornsilk             = color.RGBA{0xFF, 0xF8, 0xDC, 0xFF}
	Crimson              = color.RGBA{0xDC, 0x14, 0x3C, 0xFF}
	Cyan                 = color.RGBA{0x00, 0xFF, 0xFF, 0xFF}
	DarkBlue             = color.RGBA{0x00, 0x00, 0x8B, 0xFF}
	DarkCyan             = color.RGBA{0x00, 0x8B, 0x8B, 0xFF}
	DarkGoldenRod        = color.RGBA{0xB8, 0x86, 0x0B, 0xFF}
	DarkGray             = color.RGBA{0xA9, 0xA9, 0xA9, 0xFF}
	DarkGreen            = color.RGBA{0x00, 0x64, 0x00, 0xFF}
	DarkKhaki            = color.RGBA{0xBD, 0xB7, 0x6B, 0xFF}
	DarkMagenta          = color.RGBA{0x8B, 0x00, 0x8B, 0xFF}
	DarkOliveGreen       = color.RGBA{0x55, 0x6B, 0x2F, 0xFF}
	DarkOrange           = color.RGBA{0xFF, 0x8C, 0x00, 0xFF}
	DarkOrchid           = color.RGBA{0x99, 0x32, 0xCC, 0xFF}
	DarkRed              = color.RGBA{0x8B, 0x00, 0x00, 0xFF}
	DarkSalmon           = color.RGBA{0xE9, 0x96, 0x7A, 0xFF}
	DarkSeaGreen         = color.RGBA{0x8F, 0xBC, 0x8F, 0xFF}
	DarkSlateBlue        = color.RGBA{0x48, 0x3D, 0x8B, 0xFF}
	DarkSlateGray        = color.RGBA{0x2F, 0x4F, 0x4F, 0xFF}
	DarkTurquoise        = color.RGBA{0x2F, 0x4F, 0x4F, 0xFF}
	DarkViolet           = color.RGBA{0x94, 0x00, 0xD3, 0xFF}
	DeepPink             = color.RGBA{0xFF, 0x14, 0x93, 0xFF}
	DeepSkyBlue          = color.RGBA{0x00, 0xBF, 0xFF, 0xFF}
	DimGray              = color.RGBA{0x69, 0x69, 0x69, 0xFF}
	DodgerBlue           = color.RGBA{0x1E, 0x90, 0xFF, 0xFF}
	FireBrick            = color.RGBA{0xB2, 0x22, 0x22, 0xFF}
	FloralWhite          = color.RGBA{0xFF, 0xFA, 0xF0, 0xFF}
	ForestGreen          = color.RGBA{0x22, 0x8B, 0x22, 0xFF}
	Fuchsia              = color.RGBA{0xFF, 0x00, 0xFF, 0xFF}
	Gainsboro            = color.RGBA{0xDC, 0xDC, 0xDC, 0xFF}
	GhostWhite           = color.RGBA{0xF8, 0xF8, 0xFF, 0xFF}
	Gold                 = color.RGBA{0xFF, 0xD7, 0x00, 0xFF}
	Goldenrod            = color.RGBA{0xDA, 0xA5, 0x20, 0xFF}
	Gray                 = color.RGBA{0x80, 0x80, 0x80, 0xFF}
	Green                = color.RGBA{0x00, 0x80, 0x00, 0xFF}
	GreenYellow          = color.RGBA{0xAD, 0xFF, 0x2F, 0xFF}
	Honeydew             = color.RGBA{0xF0, 0xFF, 0xF0, 0xFF}
	HotPink              = color.RGBA{0xFF, 0x69, 0xB4, 0xFF}
	IndianRed            = color.RGBA{0xCD, 0x5C, 0x5C, 0xFF}
	Indigo               = color.RGBA{0x4B, 0x00, 0x82, 0xFF}
	Ivory                = color.RGBA{0xFF, 0xFF, 0xF0, 0xFF}
	Khaki                = color.RGBA{0xF0, 0xE6, 0x8C, 0xFF}
	Lavender             = color.RGBA{0xE6, 0xE6, 0xFA, 0xFF}
	LavenderBlush        = color.RGBA{0xFF, 0xF0, 0xF5, 0xFF}
	LawnGreen            = color.RGBA{0x7C, 0xFC, 0x00, 0xFF}
	LemonChiffon         = color.RGBA{0xFF, 0xFA, 0xCD, 0xFF}
	LightBlue            = color.RGBA{0xAD, 0xD8, 0xE6, 0xFF}
	LightCoral           = color.RGBA{0xF0, 0x80, 0x80, 0xFF}
	LightCyan            = color.RGBA{0xE0, 0xFF, 0xFF, 0xFF}
	LightGoldenrodYellow = color.RGBA{0xFA, 0xFA, 0xD2, 0xFF}
	LightGreen           = color.RGBA{0x90, 0xEE, 0x90, 0xFF}
	LightGrey            = color.RGBA{0xD3, 0xD3, 0xD3, 0xFF}
	LightPink            = color.RGBA{0xFF, 0xB6, 0xC1, 0xFF}
	LightSalmon          = color.RGBA{0xFF, 0xA0, 0x7A, 0xFF}
	LightSeaGreen        = color.RGBA{0x20, 0xB2, 0xAA, 0xFF}
	LightSkyBlue         = color.RGBA{0x87, 0xCE, 0xFA, 0xFF}
	LightSlateGray       = color.RGBA{0x77, 0x88, 0x99, 0xFF}
	LightSteelBlue       = color.RGBA{0xB0, 0xC4, 0xDE, 0xFF}
	LightYellow          = color.RGBA{0xFF, 0xFF, 0xE0, 0xFF}
	Lime                 = color.RGBA{0x00, 0xFF, 0x00, 0xFF}
	LimeGreen            = color.RGBA{0x32, 0xCD, 0x32, 0xFF}
	Linen                = color.RGBA{0xFA, 0xF0, 0xE6, 0xFF}
	Magenta              = color.RGBA{0xFF, 0x00, 0xFF, 0xFF}
	Maroon               = color.RGBA{0x80, 0x00, 0x00, 0xFF}
	MediumAquamarine     = color.RGBA{0x66, 0xCD, 0xAA, 0xFF}
	MediumBlue           = color.RGBA{0x00, 0x00, 0xCD, 0xFF}
	MediumOrchid         = color.RGBA{0xBA, 0x55, 0xD3, 0xFF}
	MediumPurple         = color.RGBA{0x93, 0x70, 0xD8, 0xFF}
	MediumSeaGreen       = color.RGBA{0x3C, 0xB3, 0x71, 0xFF}
	MediumSlateBlue      = color.RGBA{0x7B, 0x68, 0xEE, 0xFF}
	MediumSpringGreen    = color.RGBA{0x00, 0xFA, 0x9A, 0xFF}
	MediumTurquoise      = color.RGBA{0x48, 0xD1, 0xCC, 0xFF}
	MediumVioletRed      = color.RGBA{0xC7, 0x15, 0x85, 0xFF}
	MidnightBlue         = color.RGBA{0x19, 0x19, 0x70, 0xFF}
	MintCream            = color.RGBA{0xF5, 0xFF, 0xFA, 0xFF}
	MistyRose            = color.RGBA{0xFF, 0xE4, 0xE1, 0xFF}
	Moccasin             = color.RGBA{0xFF, 0xE4, 0xB5, 0xFF}
	NavajoWhite          = color.RGBA{0xFF, 0xDE, 0xAD, 0xFF}
	Navy                 = color.RGBA{0x00, 0x00, 0x80, 0xFF}
	OldLace              = color.RGBA{0xFD, 0xF5, 0xE6, 0xFF}
	Olive                = color.RGBA{0x80, 0x80, 0x00, 0xFF}
	OliveDrab            = color.RGBA{0x6B, 0x8E, 0x23, 0xFF}
	Orange               = color.RGBA{0xFF, 0xA5, 0x00, 0xFF}
	OrangeRed            = color.RGBA{0xFF, 0x45, 0x00, 0xFF}
	Orchid               = color.RGBA{0xDA, 0x70, 0xD6, 0xFF}
	PaleGoldenrod        = color.RGBA{0xEE, 0xE8, 0xAA, 0xFF}
	PaleGreen            = color.RGBA{0x98, 0xFB, 0x98, 0xFF}
	PaleTurquoise        = color.RGBA{0xAF, 0xEE, 0xEE, 0xFF}
	PaleVioletRed        = color.RGBA{0xD8, 0x70, 0x93, 0xFF}
	PapayaWhip           = color.RGBA{0xFF, 0xEF, 0xD5, 0xFF}
	PeachPuff            = color.RGBA{0xFF, 0xDA, 0xB9, 0xFF}
	Peru                 = color.RGBA{0xCD, 0x85, 0x3F, 0xFF}
	Pink                 = color.RGBA{0xFF, 0xC0, 0xCB, 0xFF}
	Plum                 = color.RGBA{0xDD, 0xA0, 0xDD, 0xFF}
	PowderBlue           = color.RGBA{0xB0, 0xE0, 0xE6, 0xFF}
	Purple               = color.RGBA{0x80, 0x00, 0x80, 0xFF}
	Red                  = color.RGBA{0xFF, 0x00, 0x00, 0xFF}
	RosyBrown            = color.RGBA{0xBC, 0x8F, 0x8F, 0xFF}
	RoyalBlue            = color.RGBA{0x41, 0x69, 0xE1, 0xFF}
	SaddleBrown          = color.RGBA{0x8B, 0x45, 0x13, 0xFF}
	Salmon               = color.RGBA{0xFA, 0x80, 0x72, 0xFF}
	SandyBrown           = color.RGBA{0xF4, 0xA4, 0x60, 0xFF}
	SeaGreen             = color.RGBA{0x2E, 0x8B, 0x57, 0xFF}
	Seashell             = color.RGBA{0xFF, 0xF5, 0xEE, 0xFF}
	Sienna               = color.RGBA{0xA0, 0x52, 0x2D, 0xFF}
	Silver               = color.RGBA{0xC0, 0xC0, 0xC0, 0xFF}
	SkyBlue              = color.RGBA{0x87, 0xCE, 0xEB, 0xFF}
	SlateBlue            = color.RGBA{0x6A, 0x5A, 0xCD, 0xFF}
	SlateGray            = color.RGBA{0x70, 0x80, 0x90, 0xFF}
	Snow                 = color.RGBA{0xFF, 0xFA, 0xFA, 0xFF}
	SpringGreen          = color.RGBA{0x00, 0xFF, 0x7F, 0xFF}
	SteelBlue            = color.RGBA{0x46, 0x82, 0xB4, 0xFF}
	Tan                  = color.RGBA{0xD2, 0xB4, 0x8C, 0xFF}
	Teal                 = color.RGBA{0x00, 0x80, 0x80, 0xFF}
	Thistle              = color.RGBA{0xD8, 0xBF, 0xD8, 0xFF}
	Tomato               = color.RGBA{0xFF, 0x63, 0x47, 0xFF}
	Turquoise            = color.RGBA{0x40, 0xE0, 0xD0, 0xFF}
	Violet               = color.RGBA{0xEE, 0x82, 0xEE, 0xFF}
	Wheat                = color.RGBA{0xF5, 0xDE, 0xB3, 0xFF}
	White                = color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
	WhiteSmoke           = color.RGBA{0xF5, 0xF5, 0xF5, 0xFF}
	Yellow               = color.RGBA{0xFF, 0xFF, 0x00, 0xFF}
	YellowGreen          = color.RGBA{0x9A, 0xCD, 0x32, 0xFF}
)
