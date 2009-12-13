LIBS=alg common fs geom gostak libtcod mem num sdl txt
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

%-run: %-cmd
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
alg-lib: common-lib mem-lib
geom-lib: num-lib common-lib
libtcod-lib: console-lib
mem-lib: common-lib
