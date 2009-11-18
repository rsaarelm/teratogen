LIBS=fomalhaut tcod

GOFILES=\
	teratogen.go\

TARG=teratogen

MAINFILE=teratogen.go

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
	cd $* && make install

%_clean:
	cd $* && make clean

%_test:
	cd $* && make test

%_nuke:
	cd $* && make nuke

include $(GOROOT)/src/Make.cmd

clean: $(LIBS_CLEAN)

nuke: $(LIBS_NUKE)