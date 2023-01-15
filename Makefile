GOFMT_FILES?=$$(find . -not -path "./vendor/*" -type f -name '*.go')

BUILD_DIR=./build
BIN_NAME=audiotheker

default: run

release-linux: fmtcheck build-setup bin-linux
	tar -czvf $(BUILD_DIR)/$(BIN_NAME).linux-amd64.tar.gz $(BUILD_DIR)/$(BIN_NAME);

release-darwin: fmtcheck build-setup bin-darwin
	tar -czvf $(BUILD_DIR)/$(BIN_NAME).darwin-amd64.tar.gz $(BUILD_DIR)/$(BIN_NAME);

release-windows: fmtcheck build-setup bin-windows
	zip -9 -y $(BUILD_DIR)/$(BIN_NAME).windows-amd64.zip $(BUILD_DIR)/$(BIN_NAME).exe;

bin-linux: fmtcheck build-setup
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BIN_NAME) -ldflags="-s -w" main.go

bin-darwin: fmtcheck build-setup
	GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BIN_NAME) -ldflags="-s -w" main.go

bin-windows: fmtcheck build-setup
	GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -o $(BUILD_DIR)/$(BIN_NAME).exe -ldflags="-s -w" main.go

bin: fmtcheck build-setup
	CGO_ENABLED=0 go build -o $(BUILD_DIR)/audiotheker -ldflags="-s -w" main.go

build-setup:
	mkdir -p $(BUILD_DIR)

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	gofmt -l $(GOFMT_FILES)

run: bin
	./audiotheker


.PHONY: bin bin-windows bin-darwin bin-linux fmt fmtcheck run default
