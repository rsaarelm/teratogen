CMDS=databake palsort sdltest teratogen

TARG=teratogen

SUB=$(LIBS:%=pkg/%) $(CMDS:%=cmd/%)

all: build.cmds

run: $(TARG).run

build.cmds: $(addsuffix .build, $(CMDS))
clean.cmds: $(addsuffix .clean, $(CMDS))

build.libs:
	$(MAKE) -C pkg all

test:
	$(MAKE) -C pkg test

# XXX: Hardwired to clean the command before build to hack around problems
# specifying library dependencies to the command.
%.run: build.libs
	$(MAKE) -C cmd/$* clean
	$(MAKE) -C cmd/$* all
	(cd ./cmd/$*; $*)

%.build: build.libs
	$(MAKE) -C cmd/$*

%.clean:
	$(MAKE) -C cmd/$* clean

clean: clean.cmds
	$(MAKE) -C pkg clean

nuke: clean.cmds
	$(MAKE) -C pkg nuke

deps:
	$(MAKE) -C pkg deps
