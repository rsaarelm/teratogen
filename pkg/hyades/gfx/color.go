package gfx

import (
	"encoding/hex"
	"fmt"
	"hyades/dbg"
	"image"
	"os"
	"regexp"
)

func StringRGB(col image.Color) string {
	if col == nil {
		return "<nil>"
	}
	r, g, b, _ := col.RGBA()
	return fmt.Sprintf("#%02x%02x%02x",
		byte(r>>24),
		byte(g>>24),
		byte(b>>24))
}

func StringRGBA(col image.Color) string {
	if col == nil {
		return "<nil>"
	}
	r, g, b, a := col.RGBA()
	return fmt.Sprintf("#%02x%02x%02x%02x",
		byte(r>>24),
		byte(g>>24),
		byte(b>>24),
		byte(a>>24))
}

func ParseColor(desc string) (col image.Color, err os.Error) {
	translated, ok := namedColors[desc]
	if ok {
		desc = translated
	}

	if !reHex.MatchString(desc) {
		err = os.NewError(fmt.Sprintf("Invalid color description '%s'", desc))
		return
	}

	var r, g, b, a string
	switch {
	// One hex digit per channel. Double the digit for the resulting color.
	case len(desc) == 4:
		r, g, b = desc[1:2]+desc[1:2], desc[2:3]+desc[2:3], desc[3:4]+desc[3:4]
		a = "ff"
	case len(desc) == 5:
		r, g, b = desc[1:2]+desc[1:2], desc[2:3]+desc[2:3], desc[3:4]+desc[3:4]
		a = desc[4:5] + desc[4:5]
	case len(desc) == 7:
		r, g, b = desc[1:3], desc[3:5], desc[5:7]
		a = "ff"
	case len(desc) == 9:
		r, g, b, a = desc[1:3], desc[3:5], desc[5:7], desc[7:9]
	default:
		err = os.NewError(fmt.Sprintf("Invalid color description '%s'", desc))
		return
	}

	col = image.RGBAColor{decodeHexByte(r), decodeHexByte(g),
		decodeHexByte(b), decodeHexByte(a),
	}
	return
}

func decodeHexByte(hexStr string) byte {
	data, err := hex.DecodeString(hexStr)
	dbg.AssertNil(err, "Error decoding hex data '%s': %s", hexStr, err)
	dbg.Assert(len(data) == 1, "Decoded unexpected amount of hex data.")
	return data[0]
}

var reHex = regexp.MustCompile("^#[a-fA-F0-9]+$")

var namedColors = map[string]string{
	"AliceBlue": "#F0F8FF",
	"AntiqueWhite": "#FAEBD7",
	"Aqua": "#00FFFF",
	"Aquamarine": "#7FFFD4",
	"Azure": "#F0FFFF",
	"Beige": "#F5F5DC",
	"Bisque": "#FFE4C4",
	"Black": "#000000",
	"BlanchedAlmond": "#FFEBCD",
	"Blue": "#0000FF",
	"BlueViolet": "#8A2BE2",
	"Brown": "#A52A2A",
	"BurlyWood": "#DEB887",
	"CadetBlue": "#5F9EA0",
	"Chartreuse": "#7FFF00",
	"Chocolate": "#D2691E",
	"Coral": "#FF7F50",
	"CornflowerBlue": "#6495ED",
	"Cornsilk": "#FFF8DC",
	"Crimson": "#DC143C",
	"Cyan": "#00FFFF",
	"DarkBlue": "#00008B",
	"DarkCyan": "#008B8B",
	"DarkGoldenRod": "#B8860B",
	"DarkGray": "#A9A9A9",
	"DarkGreen": "#006400",
	"DarkKhaki": "#BDB76B",
	"DarkMagenta": "#8B008B",
	"DarkOliveGreen": "#556B2F",
	"DarkOrange": "#FF8C00",
	"DarkOrchid": "#9932CC",
	"DarkRed": "#8B0000",
	"DarkSalmon": "#E9967A",
	"DarkSeaGreen": "#8FBC8F",
	"DarkSlateBlue": "#483D8B",
	"DarkSlateGray": "#2F4F4F",
	"DarkTurquoise": "#2F4F4F",
	"DarkViolet": "#9400D3",
	"DeepPink": "#FF1493",
	"DeepSkyBlue": "#00BFFF",
	"DimGray": "#696969",
	"DodgerBlue": "#1E90FF",
	"FireBrick": "#B22222",
	"FloralWhite": "#FFFAF0",
	"ForestGreen": "#228B22",
	"Fuchsia": "#FF00FF",
	"Gainsboro": "#DCDCDC",
	"GhostWhite": "#F8F8FF",
	"Gold": "#FFD700",
	"Goldenrod": "#DAA520",
	"Gray": "#808080",
	"Green": "#008000",
	"GreenYellow": "#ADFF2F",
	"Honeydew": "#F0FFF0",
	"HotPink": "#FF69B4",
	"IndianRed": "#CD5C5C",
	"Indigo": "#4B0082",
	"Ivory": "#FFFFF0",
	"Khaki": "#F0E68C",
	"Lavender": "#E6E6FA",
	"LavenderBlush": "#FFF0F5",
	"LawnGreen": "#7CFC00",
	"LemonChiffon": "#FFFACD",
	"LightBlue": "#ADD8E6",
	"LightCoral": "#F08080",
	"LightCyan": "#E0FFFF",
	"LightGoldenrodYellow": "#FAFAD2",
	"LightGreen": "#90EE90",
	"LightGrey": "#D3D3D3",
	"LightPink": "#FFB6C1",
	"LightSalmon": "#FFA07A",
	"LightSeaGreen": "#20B2AA",
	"LightSkyBlue": "#87CEFA",
	"LightSlateGray": "#778899",
	"LightSteelBlue": "#B0C4DE",
	"LightYellow": "#FFFFE0",
	"Lime": "#00FF00",
	"LimeGreen": "#32CD32",
	"Linen": "#FAF0E6",
	"Magenta": "#FF00FF",
	"Maroon": "#800000",
	"MediumAquamarine": "#66CDAA",
	"MediumBlue": "#0000CD",
	"MediumOrchid": "#BA55D3",
	"MediumPurple": "#9370D8",
	"MediumSeaGreen": "#3CB371",
	"MediumSlateBlue": "#7B68EE",
	"MediumSpringGreen": "#00FA9A",
	"MediumTurquoise": "#48D1CC",
	"MediumVioletRed": "#C71585",
	"MidnightBlue": "#191970",
	"MintCream": "#F5FFFA",
	"MistyRose": "#FFE4E1",
	"Moccasin": "#FFE4B5",
	"NavajoWhite": "#FFDEAD",
	"Navy": "#000080",
	"OldLace": "#FDF5E6",
	"Olive": "#808000",
	"OliveDrab": "#6B8E23",
	"Orange": "#FFA500",
	"OrangeRed": "#FF4500",
	"Orchid": "#DA70D6",
	"PaleGoldenrod": "#EEE8AA",
	"PaleGreen": "#98FB98",
	"PaleTurquoise": "#AFEEEE",
	"PaleVioletRed": "#D87093",
	"PapayaWhip": "#FFEFD5",
	"PeachPuff": "#FFDAB9",
	"Peru": "#CD853F",
	"Pink": "#FFC0CB",
	"Plum": "#DDA0DD",
	"PowderBlue": "#B0E0E6",
	"Purple": "#800080",
	"Red": "#FF0000",
	"RosyBrown": "#BC8F8F",
	"RoyalBlue": "#4169E1",
	"SaddleBrown": "#8B4513",
	"Salmon": "#FA8072",
	"SandyBrown": "#F4A460",
	"SeaGreen": "#2E8B57",
	"Seashell": "#FFF5EE",
	"Sienna": "#A0522D",
	"Silver": "#C0C0C0",
	"SkyBlue": "#87CEEB",
	"SlateBlue": "#6A5ACD",
	"SlateGray": "#708090",
	"Snow": "#FFFAFA",
	"SpringGreen": "#00FF7F",
	"SteelBlue": "#4682B4",
	"Tan": "#D2B48C",
	"Teal": "#008080",
	"Thistle": "#D8BFD8",
	"Tomato": "#FF6347",
	"Turquoise": "#40E0D0",
	"Violet": "#EE82EE",
	"Wheat": "#F5DEB3",
	"White": "#FFFFFF",
	"WhiteSmoke": "#F5F5F5",
	"Yellow": "#FFFF00",
	"YellowGreen": "#9ACD32",
}

var (
	AliceBlue            = image.RGBAColor{0xF0, 0xF8, 0xFF, 0xFF}
	AntiqueWhite         = image.RGBAColor{0xFA, 0xEB, 0xD7, 0xFF}
	Aqua                 = image.RGBAColor{0x00, 0xFF, 0xFF, 0xFF}
	Aquamarine           = image.RGBAColor{0x7F, 0xFF, 0xD4, 0xFF}
	Azure                = image.RGBAColor{0xF0, 0xFF, 0xFF, 0xFF}
	Beige                = image.RGBAColor{0xF5, 0xF5, 0xDC, 0xFF}
	Bisque               = image.RGBAColor{0xFF, 0xE4, 0xC4, 0xFF}
	Black                = image.RGBAColor{0x00, 0x00, 0x00, 0xFF}
	BlanchedAlmond       = image.RGBAColor{0xFF, 0xEB, 0xCD, 0xFF}
	Blue                 = image.RGBAColor{0x00, 0x00, 0xFF, 0xFF}
	BlueViolet           = image.RGBAColor{0x8A, 0x2B, 0xE2, 0xFF}
	Brown                = image.RGBAColor{0xA5, 0x2A, 0x2A, 0xFF}
	BurlyWood            = image.RGBAColor{0xDE, 0xB8, 0x87, 0xFF}
	CadetBlue            = image.RGBAColor{0x5F, 0x9E, 0xA0, 0xFF}
	Chartreuse           = image.RGBAColor{0x7F, 0xFF, 0x00, 0xFF}
	Chocolate            = image.RGBAColor{0xD2, 0x69, 0x1E, 0xFF}
	Coral                = image.RGBAColor{0xFF, 0x7F, 0x50, 0xFF}
	CornflowerBlue       = image.RGBAColor{0x64, 0x95, 0xED, 0xFF}
	Cornsilk             = image.RGBAColor{0xFF, 0xF8, 0xDC, 0xFF}
	Crimson              = image.RGBAColor{0xDC, 0x14, 0x3C, 0xFF}
	Cyan                 = image.RGBAColor{0x00, 0xFF, 0xFF, 0xFF}
	DarkBlue             = image.RGBAColor{0x00, 0x00, 0x8B, 0xFF}
	DarkCyan             = image.RGBAColor{0x00, 0x8B, 0x8B, 0xFF}
	DarkGoldenRod        = image.RGBAColor{0xB8, 0x86, 0x0B, 0xFF}
	DarkGray             = image.RGBAColor{0xA9, 0xA9, 0xA9, 0xFF}
	DarkGreen            = image.RGBAColor{0x00, 0x64, 0x00, 0xFF}
	DarkKhaki            = image.RGBAColor{0xBD, 0xB7, 0x6B, 0xFF}
	DarkMagenta          = image.RGBAColor{0x8B, 0x00, 0x8B, 0xFF}
	DarkOliveGreen       = image.RGBAColor{0x55, 0x6B, 0x2F, 0xFF}
	DarkOrange           = image.RGBAColor{0xFF, 0x8C, 0x00, 0xFF}
	DarkOrchid           = image.RGBAColor{0x99, 0x32, 0xCC, 0xFF}
	DarkRed              = image.RGBAColor{0x8B, 0x00, 0x00, 0xFF}
	DarkSalmon           = image.RGBAColor{0xE9, 0x96, 0x7A, 0xFF}
	DarkSeaGreen         = image.RGBAColor{0x8F, 0xBC, 0x8F, 0xFF}
	DarkSlateBlue        = image.RGBAColor{0x48, 0x3D, 0x8B, 0xFF}
	DarkSlateGray        = image.RGBAColor{0x2F, 0x4F, 0x4F, 0xFF}
	DarkTurquoise        = image.RGBAColor{0x2F, 0x4F, 0x4F, 0xFF}
	DarkViolet           = image.RGBAColor{0x94, 0x00, 0xD3, 0xFF}
	DeepPink             = image.RGBAColor{0xFF, 0x14, 0x93, 0xFF}
	DeepSkyBlue          = image.RGBAColor{0x00, 0xBF, 0xFF, 0xFF}
	DimGray              = image.RGBAColor{0x69, 0x69, 0x69, 0xFF}
	DodgerBlue           = image.RGBAColor{0x1E, 0x90, 0xFF, 0xFF}
	FireBrick            = image.RGBAColor{0xB2, 0x22, 0x22, 0xFF}
	FloralWhite          = image.RGBAColor{0xFF, 0xFA, 0xF0, 0xFF}
	ForestGreen          = image.RGBAColor{0x22, 0x8B, 0x22, 0xFF}
	Fuchsia              = image.RGBAColor{0xFF, 0x00, 0xFF, 0xFF}
	Gainsboro            = image.RGBAColor{0xDC, 0xDC, 0xDC, 0xFF}
	GhostWhite           = image.RGBAColor{0xF8, 0xF8, 0xFF, 0xFF}
	Gold                 = image.RGBAColor{0xFF, 0xD7, 0x00, 0xFF}
	Goldenrod            = image.RGBAColor{0xDA, 0xA5, 0x20, 0xFF}
	Gray                 = image.RGBAColor{0x80, 0x80, 0x80, 0xFF}
	Green                = image.RGBAColor{0x00, 0x80, 0x00, 0xFF}
	GreenYellow          = image.RGBAColor{0xAD, 0xFF, 0x2F, 0xFF}
	Honeydew             = image.RGBAColor{0xF0, 0xFF, 0xF0, 0xFF}
	HotPink              = image.RGBAColor{0xFF, 0x69, 0xB4, 0xFF}
	IndianRed            = image.RGBAColor{0xCD, 0x5C, 0x5C, 0xFF}
	Indigo               = image.RGBAColor{0x4B, 0x00, 0x82, 0xFF}
	Ivory                = image.RGBAColor{0xFF, 0xFF, 0xF0, 0xFF}
	Khaki                = image.RGBAColor{0xF0, 0xE6, 0x8C, 0xFF}
	Lavender             = image.RGBAColor{0xE6, 0xE6, 0xFA, 0xFF}
	LavenderBlush        = image.RGBAColor{0xFF, 0xF0, 0xF5, 0xFF}
	LawnGreen            = image.RGBAColor{0x7C, 0xFC, 0x00, 0xFF}
	LemonChiffon         = image.RGBAColor{0xFF, 0xFA, 0xCD, 0xFF}
	LightBlue            = image.RGBAColor{0xAD, 0xD8, 0xE6, 0xFF}
	LightCoral           = image.RGBAColor{0xF0, 0x80, 0x80, 0xFF}
	LightCyan            = image.RGBAColor{0xE0, 0xFF, 0xFF, 0xFF}
	LightGoldenrodYellow = image.RGBAColor{0xFA, 0xFA, 0xD2, 0xFF}
	LightGreen           = image.RGBAColor{0x90, 0xEE, 0x90, 0xFF}
	LightGrey            = image.RGBAColor{0xD3, 0xD3, 0xD3, 0xFF}
	LightPink            = image.RGBAColor{0xFF, 0xB6, 0xC1, 0xFF}
	LightSalmon          = image.RGBAColor{0xFF, 0xA0, 0x7A, 0xFF}
	LightSeaGreen        = image.RGBAColor{0x20, 0xB2, 0xAA, 0xFF}
	LightSkyBlue         = image.RGBAColor{0x87, 0xCE, 0xFA, 0xFF}
	LightSlateGray       = image.RGBAColor{0x77, 0x88, 0x99, 0xFF}
	LightSteelBlue       = image.RGBAColor{0xB0, 0xC4, 0xDE, 0xFF}
	LightYellow          = image.RGBAColor{0xFF, 0xFF, 0xE0, 0xFF}
	Lime                 = image.RGBAColor{0x00, 0xFF, 0x00, 0xFF}
	LimeGreen            = image.RGBAColor{0x32, 0xCD, 0x32, 0xFF}
	Linen                = image.RGBAColor{0xFA, 0xF0, 0xE6, 0xFF}
	Magenta              = image.RGBAColor{0xFF, 0x00, 0xFF, 0xFF}
	Maroon               = image.RGBAColor{0x80, 0x00, 0x00, 0xFF}
	MediumAquamarine     = image.RGBAColor{0x66, 0xCD, 0xAA, 0xFF}
	MediumBlue           = image.RGBAColor{0x00, 0x00, 0xCD, 0xFF}
	MediumOrchid         = image.RGBAColor{0xBA, 0x55, 0xD3, 0xFF}
	MediumPurple         = image.RGBAColor{0x93, 0x70, 0xD8, 0xFF}
	MediumSeaGreen       = image.RGBAColor{0x3C, 0xB3, 0x71, 0xFF}
	MediumSlateBlue      = image.RGBAColor{0x7B, 0x68, 0xEE, 0xFF}
	MediumSpringGreen    = image.RGBAColor{0x00, 0xFA, 0x9A, 0xFF}
	MediumTurquoise      = image.RGBAColor{0x48, 0xD1, 0xCC, 0xFF}
	MediumVioletRed      = image.RGBAColor{0xC7, 0x15, 0x85, 0xFF}
	MidnightBlue         = image.RGBAColor{0x19, 0x19, 0x70, 0xFF}
	MintCream            = image.RGBAColor{0xF5, 0xFF, 0xFA, 0xFF}
	MistyRose            = image.RGBAColor{0xFF, 0xE4, 0xE1, 0xFF}
	Moccasin             = image.RGBAColor{0xFF, 0xE4, 0xB5, 0xFF}
	NavajoWhite          = image.RGBAColor{0xFF, 0xDE, 0xAD, 0xFF}
	Navy                 = image.RGBAColor{0x00, 0x00, 0x80, 0xFF}
	OldLace              = image.RGBAColor{0xFD, 0xF5, 0xE6, 0xFF}
	Olive                = image.RGBAColor{0x80, 0x80, 0x00, 0xFF}
	OliveDrab            = image.RGBAColor{0x6B, 0x8E, 0x23, 0xFF}
	Orange               = image.RGBAColor{0xFF, 0xA5, 0x00, 0xFF}
	OrangeRed            = image.RGBAColor{0xFF, 0x45, 0x00, 0xFF}
	Orchid               = image.RGBAColor{0xDA, 0x70, 0xD6, 0xFF}
	PaleGoldenrod        = image.RGBAColor{0xEE, 0xE8, 0xAA, 0xFF}
	PaleGreen            = image.RGBAColor{0x98, 0xFB, 0x98, 0xFF}
	PaleTurquoise        = image.RGBAColor{0xAF, 0xEE, 0xEE, 0xFF}
	PaleVioletRed        = image.RGBAColor{0xD8, 0x70, 0x93, 0xFF}
	PapayaWhip           = image.RGBAColor{0xFF, 0xEF, 0xD5, 0xFF}
	PeachPuff            = image.RGBAColor{0xFF, 0xDA, 0xB9, 0xFF}
	Peru                 = image.RGBAColor{0xCD, 0x85, 0x3F, 0xFF}
	Pink                 = image.RGBAColor{0xFF, 0xC0, 0xCB, 0xFF}
	Plum                 = image.RGBAColor{0xDD, 0xA0, 0xDD, 0xFF}
	PowderBlue           = image.RGBAColor{0xB0, 0xE0, 0xE6, 0xFF}
	Purple               = image.RGBAColor{0x80, 0x00, 0x80, 0xFF}
	Red                  = image.RGBAColor{0xFF, 0x00, 0x00, 0xFF}
	RosyBrown            = image.RGBAColor{0xBC, 0x8F, 0x8F, 0xFF}
	RoyalBlue            = image.RGBAColor{0x41, 0x69, 0xE1, 0xFF}
	SaddleBrown          = image.RGBAColor{0x8B, 0x45, 0x13, 0xFF}
	Salmon               = image.RGBAColor{0xFA, 0x80, 0x72, 0xFF}
	SandyBrown           = image.RGBAColor{0xF4, 0xA4, 0x60, 0xFF}
	SeaGreen             = image.RGBAColor{0x2E, 0x8B, 0x57, 0xFF}
	Seashell             = image.RGBAColor{0xFF, 0xF5, 0xEE, 0xFF}
	Sienna               = image.RGBAColor{0xA0, 0x52, 0x2D, 0xFF}
	Silver               = image.RGBAColor{0xC0, 0xC0, 0xC0, 0xFF}
	SkyBlue              = image.RGBAColor{0x87, 0xCE, 0xEB, 0xFF}
	SlateBlue            = image.RGBAColor{0x6A, 0x5A, 0xCD, 0xFF}
	SlateGray            = image.RGBAColor{0x70, 0x80, 0x90, 0xFF}
	Snow                 = image.RGBAColor{0xFF, 0xFA, 0xFA, 0xFF}
	SpringGreen          = image.RGBAColor{0x00, 0xFF, 0x7F, 0xFF}
	SteelBlue            = image.RGBAColor{0x46, 0x82, 0xB4, 0xFF}
	Tan                  = image.RGBAColor{0xD2, 0xB4, 0x8C, 0xFF}
	Teal                 = image.RGBAColor{0x00, 0x80, 0x80, 0xFF}
	Thistle              = image.RGBAColor{0xD8, 0xBF, 0xD8, 0xFF}
	Tomato               = image.RGBAColor{0xFF, 0x63, 0x47, 0xFF}
	Turquoise            = image.RGBAColor{0x40, 0xE0, 0xD0, 0xFF}
	Violet               = image.RGBAColor{0xEE, 0x82, 0xEE, 0xFF}
	Wheat                = image.RGBAColor{0xF5, 0xDE, 0xB3, 0xFF}
	White                = image.RGBAColor{0xFF, 0xFF, 0xFF, 0xFF}
	WhiteSmoke           = image.RGBAColor{0xF5, 0xF5, 0xF5, 0xFF}
	Yellow               = image.RGBAColor{0xFF, 0xFF, 0x00, 0xFF}
	YellowGreen          = image.RGBAColor{0x9A, 0xCD, 0x32, 0xFF}
)
