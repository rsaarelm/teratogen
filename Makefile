GOPATH := $(CURDIR)

teratogen:
	go build teratogen

run: teratogen
	./teratogen

clean:
	go clean

.PHONY: teratogen run clean
