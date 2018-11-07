GOBUILDFLAGS=
GC=go build
SRC=main.go
PROG=streammon
cdir:=$(shell pwd)

streammon: $(SRC)
	go get -t ./...
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GC) $(GOBUILDFLAGS) -o $(PROG) $(SRC)
	chmod +x $(PROG)

test: streammon
	go test ./...
	cd tests
	tests/integration.sh $(cdir)
	cd -

clean:
	@if [ -f streammon ]; then rm streammon; fi
