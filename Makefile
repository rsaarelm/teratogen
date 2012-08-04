GOPATH := $(CURDIR)

PKGS = teratogen \
	teratogen/archive \
	teratogen/babble \
	teratogen/cache \
	teratogen/font \
	teratogen/fov \
	teratogen/gfx \
	teratogen/manifold \
	teratogen/mapgen \
	teratogen/num \
	teratogen/sdl \
	teratogen/tile \
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

clean:
	go clean

.PHONY: teratogen run clean test
