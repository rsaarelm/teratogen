GOPATH := $(CURDIR)

PKGS = teratogen \
	teratogen/action \
	teratogen/app \
	teratogen/archive \
	teratogen/babble \
	teratogen/cache \
	teratogen/display/anim \
	teratogen/display/fx \
	teratogen/display/hud \
	teratogen/display/util \
	teratogen/display/view \
	teratogen/entity \
	teratogen/factory \
	teratogen/font \
	teratogen/fov \
	teratogen/gfx \
	teratogen/kernel \
	teratogen/mapgen \
	teratogen/mob \
	teratogen/music/mod \
	teratogen/num \
	teratogen/query \
	teratogen/screen \
	teratogen/sdl \
	teratogen/ser \
	teratogen/space \
	teratogen/tile \
	teratogen/typography \
	teratogen/world \

# Zip file indexing of the zip catenated to the binary must be fixed with 'zip
# -A' so that Go's zip library will read it.
teratogen:
	go build teratogen
	strip teratogen
	rm -f assets.zip
	zip -r assets.zip assets/
	cat assets.zip >> teratogen
	zip -A teratogen

test:
	go test $(PKGS)

benchmark:
	go test -test.bench '.*' $(PKGS)

fmt:
	go fmt $(PKGS)

run: teratogen
	./teratogen

SERVERPORT=6060

doc:
	@echo "Documentation server now running at http://localhost:$(SERVERPORT)/"
	godoc -http=":$(SERVERPORT)"

clean:
	go clean

.PHONY: teratogen run clean test
