include $(GOROOT)/src/Make.$(GOARCH)

TARG=teratogen_base
GOFILES=\
	anim.go\
	config.go\
	main.go\
	mapview.go\
	media.go\
	msg.go\
	ui.go\

include $(GOROOT)/src/Make.cmd

MEDIAFILES=$(wildcard media/*)

media.tar.gz: $(MEDIAFILES)
	tar czfv $@ media

CLEANFILES+=media.tar.gz teratogen

teratogen: teratogen_base media.tar.gz
	cat teratogen_base media.tar.gz > teratogen
	chmod a+x teratogen

all: teratogen