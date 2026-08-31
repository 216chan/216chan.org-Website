FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY backend/go.mod ./
RUN go mod download
COPY backend/ .
RUN go build -o /server ./cmd/server

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /server .
EXPOSE 8080
CMD ["./server"]

