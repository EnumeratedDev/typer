# Installation paths
PREFIX ?= /usr/local
BINDIR ?= $(PREFIX)/bin
SYSCONFDIR := $(PREFIX)/etc

# Compilers and tools
GO ?= go

# Compiler flags
GOOS ?= $(shell $(GO) env | grep '^GOOS' | cut -d'=' -f2 | tr -d "'")
GOARCH ?= $(shell $(GO) env | grep '^GOARCH' | cut -d'=' -f2 | tr -d "'")
LDFLAGS ?= -w

build:
	install -dm755 build
	cd src/; GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags "$(LDFLAGS) -X 'main.sysconfdir=$(SYSCONFDIR)'" -o ../build/

install: build/typer
	# Create directories
	install -dm755 $(DESTDIR)$(BINDIR)
	# Install files
	install -m755 build/typer* $(DESTDIR)$(BINDIR)/

install-config:
	# Create directories
	install -dm755 $(DESTDIR)$(SYSCONFDIR)
	# Install files
	cp -r config -T $(DESTDIR)$(SYSCONFDIR)/typer

uninstall:
	-rm -f $(DESTDIR)$(BINDIR)/typer
	-rm -rf $(DESTDIR)$(SYSCONFDIR)/typer

clean:
	-rm -rf build/

.PHONY: build install install-config uninstall clean
