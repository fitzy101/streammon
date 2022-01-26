GOBUILDFLAGS=
GC=go build
SRC=cmd/streammon/main.go
PROG=streammon
DIST=_dist

default: $(SRC) test
	$(GC) $(GOBUILDFLAGS) -o $(DIST)/$(PROG) $(SRC)
	chmod +x $(PROG)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

cover: test
	go tool cover -html coverage.txt

clean:
	@if [ -f streammon ]; then rm streammon; fi
	@if [ -f coverage.txt ]; then rm coverage.txt; fi

