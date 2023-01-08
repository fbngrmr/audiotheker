GOFMT_FILES?=$$(find . -not -path "./vendor/*" -type f -name '*.go')

default: run

bin: fmtcheck
	CGO_ENABLED=0 go build -o audiotheker -ldflags="-s -w" main.go

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	gofmt -l $(GOFMT_FILES)

run: bin
	./audiotheker

.PHONY: bin fmt fmtcheck run default
