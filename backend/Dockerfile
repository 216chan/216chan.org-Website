FROM docker.io/library/golang:1.27-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM docker.io/library/alpine:3.22
RUN apk add --no-cache ca-certificates bubblewrap util-linux imagemagick imagemagick-jpeg imagemagick-webp \
 && adduser -D -u 10001 chan
WORKDIR /app
COPY --from=builder /server ./server
USER chan
ENV PUBLICATION_LISTEN=0.0.0.0:8080
EXPOSE 8080
CMD ["./server"]
