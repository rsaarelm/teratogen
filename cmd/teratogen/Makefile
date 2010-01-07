include $(GOROOT)/src/Make.$(GOARCH)

TARG=teratogen_base
GOFILES=\
	ai.go\
	anim.go\
	config.go\
	creature.go\
	entity.go\
	input.go\
	mapgen.go\
	main.go\
	mapview.go\
	media.go\
	msg.go\
	rules.go\
	ui.go\
	world.go\

include $(GOROOT)/src/Make.cmd

MEDIAFILES=$(wildcard media/*)

media.tar.gz: $(MEDIAFILES)
	tar czfv $@ media

CLEANFILES+=media.tar.gz teratogen

teratogen: teratogen_base media.tar.gz
	cat teratogen_base media.tar.gz > teratogen
	chmod a+x teratogen

all: teratogen