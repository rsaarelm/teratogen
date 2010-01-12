include $(GOROOT)/src/Make.$(GOARCH)

CGO_CFLAGS = $(shell sdl-config --cflags)

CGO_LDFLAGS = $(shell sdl-config --libs) -lSDL_mixer

TARG=hyades/sdl

CGOFILES=\
	sdl.go\

include $(GOROOT)/src/Make.pkg
