FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o dominator cmd/server/main.go

FROM alpine:3.19

RUN apk --no-cache add ca-certificates wget

WORKDIR /app

COPY --from=builder /app/dominator /usr/local/bin/dominator

EXPOSE 8090

ENTRYPOINT ["/usr/local/bin/dominator"]
