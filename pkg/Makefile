all: install

DIRS=\
	gostak\
	hyades/alg\
	hyades/babble\
	hyades/dbg\
	hyades/entity\
	hyades/fs\
	hyades/geom\
	hyades/gfx\
	hyades/gui\
	hyades/keyboard\
	hyades/mem\
	hyades/num\
	hyades/sdl\
	hyades/sfx\
	hyades/txt\
	teratogen/game

NOTEST=\
	hyades/dbg\
	hyades/entity\
	hyades/fs\
	hyades/gui\
	hyades/keyboard\
	hyades/sdl\
	hyades/sfx\

TEST=$(filter-out $(NOTEST),$(DIRS))

GOMAKE=$(GOBIN)/gomake

clean.dirs: $(addsuffix .clean, $(DIRS))
install.dirs: $(addsuffix .install, $(DIRS))
nuke.dirs: $(addsuffix .nuke, $(DIRS))
test.dirs: $(addsuffix .test, $(TEST))

%.clean:
	rm -f Make.deps
	$(GOMAKE) -C $* clean

%.install: Make.deps
	$(GOMAKE) -C $* install

%.nuke:
	$(GOMAKE) -C $* nuke

%.test:
	$(GOMAKE) -C $* test

clean: clean.dirs
	rm -f ../cmd/cgo

install: install.dirs

test: test.dirs

nuke: nuke.dirs

Make.deps:
	ln -sf $(GOROOT)/src/cmd/cgo ../cmd/ 	# XXX: Hack to get around the relative reference to cgo which deps.bash introduces.
	$(GOROOT)/src/pkg/deps.bash

deps: Make.deps

-include Make.deps