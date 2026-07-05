BINARY := miniclean

.PHONY: build test vet lint clean

build:
	go build -o $(BINARY) ./cmd/miniclean

test:
	go test ./...

vet:
	go vet ./...

lint: vet
	gofmt -l .

clean:
	rm -f $(BINARY) $(BINARY).exe
