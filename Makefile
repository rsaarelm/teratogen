GOPATH := $(CURDIR)

PKGS := $(shell cd src; find * -path github.com -prune -o -name \*.go -printf "%h\n" | sort | uniq)

# Zip file indexing of the zip catenated to the binary must be fixed with 'zip
# -A' so that Go's zip library will read it.
bin/teratogen: gen-version
	go install teratogen

gen-version:
	go run src/gen-version/gen-version.go

# To make Teratogen buildable with Wine:
# - Have Go installed on the Wine drive, and have C:\go\bin in Wine's cmd
#   path.
# - Have MinGW with GCC installed on the Wine drive, and have C:\MinGW\bin in
#   Wine's cmd path.
# - Have the MinGW SDL development package's bin/, lib/ and include/
#   directories in C:\MinGW on the Wine drive.
bin/teratogen.exe: gen-version
	rm -f $@
	wine go install -ldflags -H=windowsgui teratogen

dist: bin/teratogen
	mkdir -p dist/
	cp bin/teratogen .
	strip teratogen
	rm -f assets.zip
	zip -r assets.zip assets/
	cat assets.zip >> teratogen
	zip -A teratogen
	zip dist/teratogen-$$(go run src/gen-version/gen-version.go)-$$(go env GOOS)$$(go env GOARCH).zip teratogen
	rm teratogen

windist: bin/teratogen.exe
	mkdir -p dist/
	cp bin/teratogen.exe .
	strip teratogen.exe
	rm -f assets.zip
	zip -r assets.zip assets/
	cat assets.zip >> teratogen.exe
	zip -A teratogen.exe
	cp tools/win/$$(wine go env GOARCH)/SDL.dll .
	zip dist/teratogen-$$(go run src/gen-version/gen-version.go)-$$(wine go env GOOS)$$(wine go env GOARCH).zip teratogen.exe SDL.dll
	rm teratogen.exe SDL.dll

alldist: dist windist

test:
	go test $(PKGS)

benchmark:
	go test -test.bench '.*' $(PKGS)

fmt:
	go fmt $(PKGS)

run: bin/teratogen
	./bin/teratogen

SERVERPORT=6060

doc:
	@echo "Documentation server now running at http://localhost:$(SERVERPORT)/"
	godoc -http=":$(SERVERPORT)"

clean:
	go clean
	rm -rf pkg/
	rm -rf bin/
	rm -rf dist/
	rm -f assets.zip

.PHONY: bin/teratogen gen-version bin/teratogen.exe dist windist run clean test
