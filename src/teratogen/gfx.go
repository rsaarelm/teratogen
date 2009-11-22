package teratogen

import "libtcod"

const xDrawOffset = 0
const yDrawOffset = 1

type RGB [3]byte;

func DrawCharRGB(x int, y int, c int, color RGB) {
	libtcod.SetForeColor(libtcod.MakeColor(color[0], color[1], color[2]));
	libtcod.PutChar(x + xDrawOffset, y + yDrawOffset, c, libtcod.BkgndNone);
}

func DrawChar(x int, y int, c int) {
	libtcod.PutChar(x + xDrawOffset, y + yDrawOffset, c, libtcod.BkgndNone);
}

func FlushScreen() {
	libtcod.Flush();
}