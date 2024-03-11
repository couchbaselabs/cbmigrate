setup:
	./bin/setup.sh

APP_NAME := cbmigrate
VERSION := $(shell cat plugin.version)

build: clean
	GOOS=linux GOARCH=amd64 go build -ldflags="-X 'main.Version=1.0.0'" -o builds/bin/linux_amd64/$(APP_NAME)
	GOOS=linux GOARCH=arm64 go build -ldflags="-X 'main.Version=1.0.0'" -o builds/bin/linux_arm64/$(APP_NAME)
	GOOS=darwin GOARCH=amd64 go build -ldflags="-X 'main.Version=1.0.0'" -o builds/bin/darwin_amd64/$(APP_NAME)
	GOOS=darwin GOARCH=arm64 go build -ldflags="-X 'main.Version=1.0.0'" -o builds/bin/darwin_arm64/$(APP_NAME)
	GOOS=windows GOARCH=amd64 go build -ldflags="-X 'main.Version=1.0.0'" -o builds/bin/windows_amd64/$(APP_NAME).exe
	mkdir -p -m 777 builds/zip

	zip -j builds/zip/$(APP_NAME)_$(VERSION)_linux_amd64.zip builds/bin/linux_amd64/$(APP_NAME)
	zip -j builds/zip/$(APP_NAME)_$(VERSION)_linux_arm64.zip builds/bin/linux_arm64/$(APP_NAME)
	zip -j builds/zip/$(APP_NAME)_$(VERSION)_darwin_amd64.zip builds/bin/darwin_amd64/$(APP_NAME)
	zip -j builds/zip/$(APP_NAME)_$(VERSION)_darwin_arm64.zip builds/bin/darwin_arm64/$(APP_NAME)
	zip -j builds/zip/$(APP_NAME)_$(VERSION)_windows_amd64.zip builds/bin/windows_amd64/$(APP_NAME).exe

clean:
	rm -rf builds/*

test:
	 go test ./... -coverprofile=coverage.out

fmt:
	 go fmt ./...

gog:
	go generate ./...

lint:
	golangci-lint run --config=.golangci.yml