APP_NAME := openclaw-setup
WEB_DIR := web
DIST_DIR := dist
UPX ?= upx
UPX_FLAGS ?= --best --lzma

define upx_compress
	@if command -v $(UPX) >/dev/null 2>&1; then \
		$(UPX) $(UPX_FLAGS) $(1); \
	else \
		echo "upx not found, skip compression"; \
	fi
endef

.PHONY: build build-web build-go build-linux build-linux-arm64 clean

build: build-web build-go

build-web:
	cd $(WEB_DIR) && npm install && npm run build

build-go:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(APP_NAME) ./cmd/server
	$(call upx_compress,$(DIST_DIR)/$(APP_NAME))

build-linux:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server
	$(call upx_compress,$(DIST_DIR)/$(APP_NAME)-linux-amd64)

build-linux-arm64:
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=arm64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server
	$(call upx_compress,$(DIST_DIR)/$(APP_NAME)-linux-arm64)

clean:
	rm -rf $(DIST_DIR)
