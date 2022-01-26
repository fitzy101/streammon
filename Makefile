PROG=streammon
DIST=_dist
SRC=cmd/streammon/main.go

GOBUILDFLAGS=
ARCH=$(or $(GOARCH),amd64)
OS=$(or $(GOOS),linux)
OUTPUT=$(DIST)/$(PROG)-$(OS)-$(ARCH)

default: $(SRC) test
	go build $(GOBUILDFLAGS) -o $(OUTPUT) $(SRC)
	chmod +x $(OUTPUT)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

cover: test
	go tool cover -html coverage.txt

clean:
	@if [ -f streammon ]; then rm streammon; fi
	@if [ -f coverage.txt ]; then rm coverage.txt; fi

