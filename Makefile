bin/topcolors: bin
	go build -o $@ ./cmd/topcolors/...

bin:
	mkdir -p bin

.PHONY: clean
clean:
	@rm -rf /tmp/topcolors.csv
	@rm -rf bin
	@rm -rf /tmp/topcolors.db

.PHONY: test
test:
	go test -v ./...
