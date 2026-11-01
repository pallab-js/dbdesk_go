SHELL := /bin/bash

BIN      := build/bin/dbdesk.app/Contents/MacOS/dbdesk
TAGS     := webkit2_41
WAILS    ?= $(shell command -v wails 2>/dev/null || echo $(HOME)/go/bin/wails)
DESTDIR  ?=
PREFIX   ?= /usr/local
APPDIR   := $(DESTDIR)$(PREFIX)/share/applications
ICONDIR  := $(DESTDIR)$(PREFIX)/share/icons/hicolor/1024x1024/apps

.PHONY: all build dev run test check clean install uninstall

all: build

# webkit2_41 is required on Ubuntu 24.04+/26.04 with libwebkit2gtk-4.1.
# Drop the tag on distributions that ship the 4.0 development package.
build:
	$(WAILS) build -tags $(TAGS)

dev:
	$(WAILS) dev -tags $(TAGS)

run: build
	./$(BIN)

test:
	go test ./...

check:
	cd frontend && npm run check

clean:
	rm -rf build/bin frontend/dist

install: build
	install -Dm755 $(BIN) $(DESTDIR)$(PREFIX)/bin/dbdesk
	install -Dm644 build/dbdesk.desktop $(APPDIR)/dbdesk.desktop
	install -Dm644 build/appicon.png $(ICONDIR)/dbdesk.png
	@echo "Installed dbdesk to $(DESTDIR)$(PREFIX)/bin"

uninstall:
	rm -f $(DESTDIR)$(PREFIX)/bin/dbdesk
	rm -f $(APPDIR)/dbdesk.desktop
	rm -f $(ICONDIR)/dbdesk.png
