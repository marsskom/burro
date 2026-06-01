-include .env
export

CA_CERT=runtime/certs/ca.pem
CA_NAME=Burro CA
KEYCHAIN=/Library/Keychains/System.keychain

PROXY=burro

ARGS ?=

.PHONY: certs
certs:
	go run ./cmd/burro cert $(ARGS)

.PHONY: ca-install
ca-install:
	@echo "Installing CA into macOS System Keychain..."
	@sudo security add-trusted-cert \
		-d \
		-r trustRoot \
		-k $(KEYCHAIN) \
		$(CA_CERT)
	@echo "Done. CA installed."

.PHONY: ca-remove
ca-remove:
	@echo "Removing CA from System Keychain..."
	@sudo security delete-certificate \
		-c "$(CA_NAME)" \
		$(KEYCHAIN) || true
	@echo "Done. CA removed."

.PHONY: ca-find
ca-find:
	@echo "Searching CA in System Keychain..."
	@security find-certificate -c "$(CA_NAME)" $(KEYCHAIN) || true


.PHONY: gen
gen:
	go generate ./tools/plugin-gen

.PHONY: build
build:
	$(MAKE) gen
	go build -o $(PROXY) ./cmd/burro

.PHONY: proxy
proxy:
	$(MAKE) gen
	go run ./cmd/burro -d runtime proxy $(ARGS)

.PHONY: serve
serve:
	$(MAKE) gen
	go run ./cmd/burro serve $(ARGS)

.PHONY: protoge
protoge:
	protoc \
   -I api \
   --go_out=./internal/proto/ \
   --go-grpc_out=./internal/proto/ \
   --go_opt=paths=source_relative \
   --go-grpc_opt=paths=source_relative \
   api/burro.proto

.PHONY: browser
browser:
	chromium \
		--proxy-server="http://localhost:8080" \
		--user-data-dir=/tmp/burro \
		--no-first-run \
		--no-default-browser-check \
		--disable-background-networking \
		--disable-sync \
		--disable-translate \
		--disable-component-update \
		--disable-client-side-phishing-detection \
		--disable-domain-reliability \
		--disable-features=HeavyAdIntervention,PrivacySandboxSettings3,IsolateOrigins,site-per-process \
		--metrics-recording-only \
		--safebrowsing-disable-auto-update \
		--safebrowsing-disable-download-protection

.PHONY: docker-build
docker-build:
	docker build -t $(PROJECT) .

.PHONY: docker-run
docker-run:
	docker run -it --rm -p 8080:8080 \
		-v ./runtime:/usr/src/app/runtime \
		$(PROJECT) $(ARGS)

.PHONY: toose
toose:
	GOOSE_DRIVER="sqlite3" \
		GOOSE_DBSTRING="./runtime/db/goose.sqlite3" \
		GOOSE_MIGRATION_DIR="./internal/migrations/sql/schema" \
		goose $(ARGS) 
