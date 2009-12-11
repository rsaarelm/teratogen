LIBS=gamelib libtcod sdl
CMDS=teratogen databake sdltest

TARG=teratogen

SUB=$(LIBS:%=pkg/%) $(CMDS:%=cmd/%)

LIBSUFFIX=a

LIB_BUILD:=$(LIBS:%=%-lib)
CMD_BUILD:=$(CMDS:%=%-cmd)
CMD_RUN:=$(CMDS:%=%-run)
SUB_CLEAN:=$(SUB:%=%-clean)
SUB_TEST:=$(SUB:%=%-test)
SUB_NUKE:=$(SUB:%=%-nuke)
LIB_FILES:=$(LIBS:%=$(GOROOT)/pkg/$(GOOS)_$(GOARCH)/%.$(LIBSUFFIX))

all: $(CMD_BUILD)

$(CMD_BUILD): $(LIB_BUILD)

test: $(SUB_TEST)

%-run: %-cmd
	(cd ./cmd/$*; $*)

%-lib:
	$(MAKE) -C pkg/$* install

%-cmd:
	$(MAKE) -C cmd/$*

%-clean:
	$(MAKE) -C $* clean

%-test:
	$(MAKE) -C $* test

%-nuke:
	$(MAKE) -C $* nuke

clean: $(SUB_CLEAN)

nuke: $(SUB_NUKE)