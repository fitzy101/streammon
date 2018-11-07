GOBUILDFLAGS=
GC=go build
SRC=main.go
PROG=streammon

streammon: $(SRC)
	go get -t ./...
	go test -race -coverprofile=coverage.txt -covermode=atomic
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...
	$(GC) $(GOBUILDFLAGS) -o $(PROG) $(SRC)
	chmod +x $(PROG)

clean:
	@if [ -f streammon ]; then rm streammon; fi
