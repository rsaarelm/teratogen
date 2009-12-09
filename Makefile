LIBS=gamelib gostak libtcod teratogen

GOFILES=src/main.go

TARG=teratogen

MAINFILE=src/main.go

# Start default makefile
include $(GOROOT)/src/Make.$(GOARCH)

LIBSUFFIX=a

LIBS_BUILD:=$(LIBS:%=%_build)
LIBS_CLEAN:=$(LIBS:%=%_clean)
LIBS_TEST:=$(LIBS:%=%_test)
LIBS_NUKE:=$(LIBS:%=%_nuke)
LIB_FILES:=$(LIBS:%=$(GOROOT)/pkg/$(GOOS)_$(GOARCH)/%.$(LIBSUFFIX))

all: $(TARG)

$(TARG): $(LIB_FILES)

# XXX: Hack to make the main object file get recompiled if the libraries have
# been updated.
_go_.8: $(LIB_FILES)

# XXX: Hacky dependency to the main file to ensure that the libraries get
# built before we try to compile the main file.
$(MAINFILE): $(LIBS_BUILD)

test: $(LIBS_TEST)

run: $(TARG)
	./$(TARG)

%_build:
	$(MAKE) -C src/$* install

%_clean:
	$(MAKE) -C src/$* clean

%_test:
	$(MAKE) -C src/$* test

%_nuke:
	$(MAKE) -C src/$* nuke

include $(GOROOT)/src/Make.cmd

clean: $(LIBS_CLEAN)

nuke: $(LIBS_NUKE)