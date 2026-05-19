-include .env
export

PROXY=burro-proxy

gen:
	go generate ./tools/plugin-gen

build:
	go build -o bin/$(PROXY) ./cmd/proxy

run:
	go generate ./tools/plugin-gen && BURRO_CONFIG=config.yml && go run ./cmd/proxy

docker-build:
	docker build -t $(PROJECT) .

docker-run:
	docker run --rm -p 8080:8080 $(PROJECT)
