setup:
	./bin/setup.sh

build:
	GOOS=linux GOARCH=amd64 go build -o builds/cbmigrate-linux-amd64
	GOOS=linux GOARCH=arm go build -o builds/cbmigrate-linux-arm
	GOOS=darwin GOARCH=amd64 go build -o builds/cbmigrate-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build -o builds/cbmigrate-darwin-arm64
	GOOS=windows GOARCH=amd64 go build -o builds/cbmigrate-windows-amd64.exe

test:
	 go test ./... -coverprofile=coverage.out

fmt:
	 go fmt ./...

gog:
	go generate ./...

lint:
	golangci-lint run --config=.golangci.yml