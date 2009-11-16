all: teratogen

include $(GOROOT)/src/Make.$(GOARCH)

TARG=teratogen
GOFILES=\
	teratogen.go\

run: teratogen
	./teratogen

teratogen: tcod_lib fomalhaut_lib

%_lib:
	cd $* && make install

include $(GOROOT)/src/Make.cmd
