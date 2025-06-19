FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o reporter ./cmd/reporter

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/reporter .
COPY config.yaml /root/config.yaml

CMD ["./reporter", "/root/config.yaml"]