all: deps install

DIRS=\
	hyades/alg\
	hyades/dbg\
	hyades/event\
	hyades/fs\
	hyades/geom\
	hyades/gfx\
	hyades/gostak\
	hyades/mem\
	hyades/num\
	hyades/sdl\
	hyades/sfx\
	hyades/txt\

NOTEST=\
	hyades/dbg\
	hyades/event\
	hyades/fs\
	hyades/geom\
	hyades/num\
	hyades/sdl\
	hyades/sfx\

TEST=$(filter-out $(NOTEST),$(DIRS))

clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
nuke.dirs: $(addsuffix .nuke, $(DIRS))
test.dirs: $(addsuffix .test, $(TEST))

%.clean:
	$(MAKE) -C $* clean

%.install:
	$(MAKE) -C $* install

%.nuke:
	$(MAKE) -C $* nuke

%.test:
	$(MAKE) -C $* test

clean: clean.dirs

install: install.dirs

test: test.dirs

nuke: nuke.dirs

deps:
	$(GOROOT)/src/pkg/deps.bash

-include Make.deps