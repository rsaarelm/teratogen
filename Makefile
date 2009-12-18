LIBS=alg console dbg event fs geom gfx gostak libtcod mem num sdl sfx txt
CMDS=teratogen databake sdltest

TARG=teratogen

SUB=$(LIBS:%=pkg/%) $(CMDS:%=cmd/%)

LIBSUFFIX=a

LIB_BUILD:=$(LIBS:%=%-lib)
CMD_BUILD:=$(CMDS:%=%-cmd)
CMD_RUN:=$(CMDS:%=%-run)
SUB_CLEAN:=$(SUB:%=%-clean)
LIB_TEST:=$(LIBS:%=%-test)
SUB_NUKE:=$(SUB:%=%-nuke)
LIB_FILES:=$(LIBS:%=$(GOROOT)/pkg/$(GOOS)_$(GOARCH)/%.$(LIBSUFFIX))

all: $(CMD_BUILD)

$(CMD_BUILD): $(LIB_BUILD)

test: $(LIB_TEST)

# XXX: Hardwired to clean, since library dependencies from cmds don't work
# right otherwise.
%-run: clean %-cmd
	(cd ./cmd/$*; $*)

%-lib:
	$(MAKE) -C pkg/$* install

%-cmd:
	$(MAKE) -C cmd/$*

%-clean:
	$(MAKE) -C $* clean

%-test:
	$(MAKE) -C pkg/$* test

%-nuke:
	$(MAKE) -C $* nuke

clean: $(SUB_CLEAN)

nuke: $(SUB_NUKE)

# Library interdependencies
alg-lib: dbg-lib geom-lib mem-lib
geom-lib: num-lib
gfx-lib: dbg-lib num-lib
console-lib: dbg-lib geom-lib
libtcod-lib: console-lib dbg-lib
mem-lib: dbg-lib
sdl-lib: dbg-lib event-lib
sfx-lib: dbg-lib num-lib