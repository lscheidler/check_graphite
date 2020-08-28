VERSION := $(shell grep "version =" main.go | cut -d '"' -f 2)
PROGNAME := check_graphite

fmt:
	go fmt ./ ./check ./graphite 

build: fmt
	GOOS=darwin GOARCH=amd64 go build -o build/darwin_amd64/$(PROGNAME)
	GOOS=linux GOARCH=amd64 go build -o build/linux_amd64/$(PROGNAME)

dist: clean build
	mkdir dist
	zip -j dist/$(PROGNAME)_$(VERSION)_darwin_amd64.zip build/darwin_amd64/$(PROGNAME)
	zip -j dist/$(PROGNAME)_$(VERSION)_linux_amd64.zip build/linux_amd64/$(PROGNAME)
	cd dist && sha512sum *.zip > $(PROGNAME)_$(VERSION)_SHA512SUM.txt

clean:
	rm -rf dist

sign:
	gpg --armor --sign --detach-sig dist/$(PROGNAME)_$(VERSION)_darwin_amd64.zip
	gpg --armor --sign --detach-sig dist/$(PROGNAME)_$(VERSION)_linux_amd64.zip

release:
	@echo "| File | Sign  | SHA512SUM |"
	@echo "|------|-------|-----------|"
	@echo "| [$(PROGNAME)_$(VERSION)_darwin_amd64.zip](../../releases/download/$(VERSION)/$(PROGNAME)_$(VERSION)_darwin_amd64.zip) | [$(PROGNAME)_$(VERSION)_darwin_amd64.zip.asc](../../releases/download/$(VERSION)/$(PROGNAME)_$(VERSION)_darwin_amd64.zip.asc) | $(shell sha512sum dist/$(PROGNAME)_$(VERSION)_darwin_amd64.zip | cut -d " " -f 1) |"
	@echo "| [$(PROGNAME)_$(VERSION)_linux_amd64.zip](../../releases/download/$(VERSION)/$(PROGNAME)_$(VERSION)_linux_amd64.zip) | [$(PROGNAME)_$(VERSION)_linux_amd64.zip.asc](../../releases/download/$(VERSION)/$(PROGNAME)_$(VERSION)_linux_amd64.zip.asc) | $(shell sha512sum dist/$(PROGNAME)_$(VERSION)_linux_amd64.zip | cut -d " " -f 1) |"
