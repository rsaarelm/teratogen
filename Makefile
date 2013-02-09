GOPATH := $(CURDIR)

PKGS := $(shell cd src; find * -path github.com -prune -o -name \*.go -printf "%h\n" | sort | uniq)

# Zip file indexing of the zip catenated to the binary must be fixed with 'zip
# -A' so that Go's zip library will read it.
bin/teratogen: bin/gen-version
	bin/gen-version
	go install teratogen

bin/gen-version:
	go install gen-version

# To make Teratogen buildable with Wine:
# - Have Go installed on the Wine drive, and have C:\go\bin in Wine's cmd
#   path.
# - Have MinGW with GCC installed on the Wine drive, and have C:\MinGW\bin in
#   Wine's cmd path.
# - Have the MinGW SDL development package's bin/, lib/ and include/
#   directories in C:\MinGW on the Wine drive.
bin/teratogen.exe: bin/gen-version
	rm -f $@
	bin/gen-version
	wine go install -ldflags -Hwindowsgui teratogen

dist: bin/teratogen
	mkdir -p dist/
	cp bin/teratogen .
	strip teratogen
	rm -f assets.zip
	zip -r assets.zip assets/
	cat assets.zip >> teratogen
	zip -A teratogen
	zip dist/teratogen-$$(go run src/gen-version/gen-version.go)-linux_$$(uname -m).zip teratogen
	rm teratogen

windist: bin/teratogen.exe
	mkdir -p dist/
	cp bin/teratogen.exe .
	strip teratogen.exe
	rm -f assets.zip
	zip -r assets.zip assets/
	cat assets.zip >> teratogen.exe
	zip -A teratogen.exe
	zip dist/teratogen-$$(go run src/gen-version/gen-version.go)-win32.zip teratogen.exe
	rm teratogen.exe

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

.PHONY: bin/teratogen bin/gen-version bin/teratogen.exe dist windist run clean test
