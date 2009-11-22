include $(GOROOT)/src/Make.$(GOARCH)

CGO_CFLAGS = -I$(LIBTCOD)/include

CGO_LDFLAGS = -L$(LIBTCOD) -ltcod

TARG=libtcod
CGOFILES=\
	libtcod.go

include $(GOROOT)/src/Make.pkg
