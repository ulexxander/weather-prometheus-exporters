FROM golang:1.17 AS builder
WORKDIR /build
ENV \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build ./main.go

FROM alpine
WORKDIR /weather-prometheus-exporters
COPY --from=builder /build/main .
EXPOSE 80
ENTRYPOINT ["/weather-prometheus-exporters/main"]
