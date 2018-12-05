GOBUILDFLAGS=
GC=go build
SRC=cmd/streammon.go
PROG=streammon

streammon: $(SRC) depend test
	$(GC) $(GOBUILDFLAGS) -o $(PROG) $(SRC)
	chmod +x $(PROG)

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

depend:
	go get -t ./...

cover: test
	go tool cover -html coverage.txt

clean:
	@if [ -f streammon ]; then rm streammon; fi
	@if [ -f coverage.txt ]; then rm coverage.txt; fi

