include $(GOROOT)/src/Make.inc

TARG=teratogen_base
GOFILES=\
	anim.go\
	config.go\
	main.go\
	mapview.go\
	media.go\
	msg.go\
	repl.go\
	ui.go\

include $(GOROOT)/src/Make.cmd

SOFILES=\
	$(GOROOT)/pkg/$(GOOS)_$(GOARCH)/cgo_hyades_sdl.so $(GOROOT)/pkg/$(GOOS)_$(GOARCH)/libcgo.so

# Redefine link rule to set the path for the shared objects at binary
# location. $ORIGIN is the magic rpath that always looks from where the binary
# is.
$(TARG): _go_.$O $(OFILES)
	$(LD) -r \$$ORIGIN -o $@ _go_.$O $(OFILES)

MEDIAFILES=$(wildcard media/*)

media.tar.gz: $(MEDIAFILES)
	tar czfv $@ media

CLEANFILES+=media.tar.gz teratogen cgo_hyades_sdl.so libcgo.so

teratogen: teratogen_base media.tar.gz
	cat teratogen_base media.tar.gz > teratogen
	chmod a+x teratogen
	cp $(SOFILES) .

all: teratogen