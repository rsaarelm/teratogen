all: deps install

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
	teratogen\

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
	$(GOMAKE) -C $* clean

%.install:
	$(GOMAKE) -C $* install

%.nuke:
	$(GOMAKE) -C $* nuke

%.test:
	$(GOMAKE) -C $* test

clean: clean.dirs
	rm -f Make.deps
	rm -f ../cmd/cgo

install: install.dirs

test: test.dirs

nuke: nuke.dirs

deps:
	ln -sf $(GOROOT)/src/cmd/cgo ../cmd/ 	# XXX: Hack to get around the relative reference to cgo which deps.bash introduces.
	$(GOROOT)/src/pkg/deps.bash

-include Make.deps