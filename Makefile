BINARY_NAME=envman
INSTALL_DIR=/usr/local/bin
BUILD_DIR=./bin

.PHONY: build install clean test coverage

build:
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/envman

install: build
	cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_DIR)/$(BINARY_NAME)

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./...

coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	go tool cover -func=coverage.out

lint:
	go vet ./...
