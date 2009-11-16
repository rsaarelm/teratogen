all: teratogen

include $(GOROOT)/src/Make.$(GOARCH)

CGO_CFLAGS = -I$(LIBTCOD)/include
CGO_LDFLAGS = -L$(LIBTCOD) -ltcod

# XXX: Don't like installing everything to main Go site, but that seems to be
# the way to currently do things.
FOMALHAUT_LIB = $(GOROOT)/pkg/$(GOOS)_$(GOARCH)/fomalhaut.a

TARG=tcod
CGOFILES=\
	tcod.go

CLEANFILES+=teratogen

include $(GOROOT)/src/Make.pkg

run: teratogen
	./teratogen

$(FOMALHAUT_LIB):
	cd fomalhaut && make install

teratogen: $(FOMALHAUT_LIB)

%: install %.go
	$(GC) $*.go
	$(LD) -o $@ $*.$O

