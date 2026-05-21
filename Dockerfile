FROM golang:1.26

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate ./tools/plugin-gen

RUN go run ./cmd/certgen

RUN go build -v -o /usr/local/bin/burro-proxy ./cmd/proxy

CMD ["/usr/local/bin/burro-proxy"]
