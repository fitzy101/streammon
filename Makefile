GOBUILDFLAGS=
GC=go build
SRC=main.go
PROG=streammon

streammon: $(SRC)
	go get -t ./...
	go test ./...
	$(GC) $(GOBUILDFLAGS) -o $(PROG) $(SRC)
	chmod +x $(PROG)

clean:
	@if [ -f streammon ]; then rm streammon; fi
