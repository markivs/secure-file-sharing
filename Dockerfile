FROM golang:1.22-alpine3.19

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /daemon ./cmd/daemon/main.go

ENTRYPOINT ["/daemon"]
