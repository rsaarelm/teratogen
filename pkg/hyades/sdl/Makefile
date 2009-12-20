include $(GOROOT)/src/Make.$(GOARCH)

CGO_CFLAGS = $(shell sdl-config --cflags)

CGO_LDFLAGS = $(shell sdl-config --libs) -lSDL_mixer

TARG=hyades/sdl

GOFILES=\
	sdldef.go\

CGOFILES=\
	sdl.go\

# Generate Go-friendly constants and structs.
sdldef.go: sdldef.c
	godefs -g sdl $< > $@

CLEANFILES+=sdldef.go

include $(GOROOT)/src/Make.pkg
