.PHONY: build install setup test run clean

PREFIX ?= $(HOME)/.local
BINDIR ?= $(PREFIX)/bin

build:
	mkdir -p bin
	go build -o bin/tsnip ./cmd/tsnip

install: build
	mkdir -p $(BINDIR)
	install -m 755 bin/tsnip $(BINDIR)/tsnip

# Install binary + wire Ctrl+G into ~/.zshrc or ~/.bashrc
setup:
	@bash ./install.sh

test:
	go test ./...

run: build
	./bin/tsnip

clean:
	rm -rf bin
