APP_NAME := booker
GO       := go

.PHONY: install build test clean

install:
	$(GO) install .

build:
	$(GO) build -o $(APP_NAME) .

test:
	$(GO) test ./...

clean:
	rm -f $(APP_NAME)
