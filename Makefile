LIBS=gamelib gostak libtcod
CMDS=teratogen databake

TARG=teratogen

SUB=$(LIBS:%=pkg/%) $(CMDS:%=cmd/%)

LIBSUFFIX=a

SUB_BUILD:=$(SUB:%=%_build)
SUB_CLEAN:=$(SUB:%=%_clean)
SUB_TEST:=$(SUB:%=%_test)
SUB_NUKE:=$(SUB:%=%_nuke)
LIB_FILES:=$(LIBS:%=$(GOROOT)/pkg/$(GOOS)_$(GOARCH)/%.$(LIBSUFFIX))

all: cmd/$(TARG)_build

cmd/$(TARG)_build: $(LIBS:%=pkg/%_build)

# XXX: Hack to make the main object file get recompiled if the libraries have
# been updated.
cmd/$(TARG)/_go_.8: $(LIB_FILES)

# XXX: Hacky dependency to the main file to ensure that the libraries get
# built before we try to compile the main file.
$(MAINFILE): $(SUB_BUILD)

test: $(SUB_TEST)

run: cmd/$(TARG)_build
	./cmd/$(TARG)/$(TARG)

%_build:
	$(MAKE) -C $* install

%_clean:
	$(MAKE) -C $* clean

%_test:
	$(MAKE) -C $* test

%_nuke:
	$(MAKE) -C $* nuke

clean: $(SUB_CLEAN)

nuke: $(SUB_NUKE)