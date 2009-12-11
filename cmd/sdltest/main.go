package main

import "time"

import "hyades/sdl"

func main() {
	sdl.InitSdl(640, 480, "Hello SDL", false);
	time.Sleep(2e9);
	sdl.ExitSdl();
}