GOPATH := $(CURDIR)

# Zip file indexing of the zip catenated to the binary must be fixed with 'zip
# -A' so that Go's zip library will read it.
teratogen:
	go build teratogen
	strip teratogen
	rm -f assets.zip
	zip -r assets.zip assets/
	cat assets.zip >> teratogen
	zip -A teratogen

run: teratogen
	./teratogen

clean:
	go clean

.PHONY: teratogen run clean
