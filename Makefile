LIBS=fomalhaut libtcod teratogen

# Library dependencies.
teratogen_build: fomalhaut_build libtcod_build

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

# XXX: Hacky dependency to the main file to ensure that the libraries get
# built before we try to compile the main file.
$(MAINFILE): $(LIBS_BUILD)

test: $(LIBS_TEST)

run: $(TARG)
	./$(TARG)

%_build:
	cd src/$* && make install

%_clean:
	cd src/$* && make clean

%_test:
	cd src/$* && make test

%_nuke:
	cd src/$* && make nuke

include $(GOROOT)/src/Make.cmd

clean: $(LIBS_CLEAN)

nuke: $(LIBS_NUKE)