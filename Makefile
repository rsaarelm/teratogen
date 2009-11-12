all: teratogen

include $(GOROOT)/src/Make.$(GOARCH)

CGO_CFLAGS = -I$(LIBTCOD)/include
CGO_LDFLAGS = -L$(LIBTCOD) -ltcod

TARG=tcod
CGOFILES=\
	tcod.go

CLEANFILES+=teratogen

include $(GOROOT)/src/Make.pkg

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O
