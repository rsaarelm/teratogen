LIBS=fomalhaut tcod

GOFILES=\
	teratogen.go\

TARG=teratogen

MAINFILE=teratogen.go


# Start default makefile
include $(GOROOT)/src/Make.$(GOARCH)

LIBS_BUILD:=$(LIBS:%=%_build)
LIBS_CLEAN:=$(LIBS:%=%_clean)

all: $(TARG)

# XXX: Hacky dependency to the main file to ensure that the libraries get
# built before we try to compile the main file.
$(MAINFILE): $(LIBS_BUILD)

run: $(TARG)
	./$(TARG)

%_build:
	cd $* && make install

%_clean:
	cd $* && make clean

include $(GOROOT)/src/Make.cmd

clean: $(LIBS_CLEAN)