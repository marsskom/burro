-include .env
export

PROXY=burro-proxy

build:
	go build -o bin/$(PROXY) ./cmd/proxy

run:
	go run ./cmd/proxy

docker-build:
	docker build -t $(PROJECT) .

docker-run:
	docker run --rm -p 8080:8080 $(PROJECT)
