APP_NAME := openclaw-setup
WEB_DIR := web
DIST_DIR := dist

.PHONY: build build-web build-go build-linux build-linux-arm64 clean

build: build-web build-go

build-web:
	cd $(WEB_DIR) && npm install && npm run build

build-go:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(APP_NAME) ./cmd/server

build-linux:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server

build-linux-arm64:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server

clean:
	rm -rf $(DIST_DIR)
