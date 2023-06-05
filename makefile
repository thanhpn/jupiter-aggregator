.PHONY: dev
dev:
	go run main.go

.PHONY: install
install:
	go mod download
