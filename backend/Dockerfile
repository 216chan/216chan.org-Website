FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -o /server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /server .
COPY --from=builder /app/migrations ./migrations
RUN mkdir -p /data/uploads
EXPOSE 8080
CMD ["./server"]


