APP_NAME := booker
GO       := go

.PHONY: install build test clean lint vet fmt cover evolve evolve-local cron-install cron-uninstall cron-status

install:
	$(GO) install .

build:
	$(GO) build -o $(APP_NAME) .

test:
	$(GO) test ./...

clean:
	rm -f $(APP_NAME) cover.out

lint:
	golangci-lint run

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .

cover:
	$(GO) test -coverprofile=cover.out ./... && $(GO) tool cover -func=cover.out

evolve:
	bash scripts/evolve.sh

evolve-local:
	bash scripts/evolve-local.sh

cron-install:
	bash scripts/install-cron.sh install

cron-uninstall:
	bash scripts/install-cron.sh uninstall

cron-status:
	bash scripts/install-cron.sh status
