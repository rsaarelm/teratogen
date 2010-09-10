include $(GOROOT)/src/Make.inc

CGO_CFLAGS = $(shell sdl-config --cflags)

CGO_LDFLAGS = $(shell sdl-config --libs) -lSDL_ttf -lSDL_gfx

TARG=hyades/sdl

CGOFILES=\
	sdl.go\

include $(GOROOT)/src/Make.pkg
