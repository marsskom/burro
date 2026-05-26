FROM golang:1.26 AS builder

RUN go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1
RUN go install github.com/pressly/goose/v3/cmd/goose@v3.27.1 

FROM golang:1.26

WORKDIR /usr/src/app

COPY --from=builder /go/bin/sqlc /usr/local/bin/sqlc

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate ./tools/plugin-gen

RUN go run ./cmd/certgen

RUN go build -v -o /usr/local/bin/burro-proxy ./cmd/proxy

ENTRYPOINT ["/usr/local/bin/burro-proxy"]
