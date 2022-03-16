FROM golang:1.17 AS builder
WORKDIR /build
ENV \
  GO111MODULE=on \
  CGO_ENABLED=0 \
  GOOS=linux \
  GOARCH=amd64
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build ./main.go

FROM alpine
WORKDIR /open-weather-prometheus-exporter
COPY --from=builder /build/main .
EXPOSE 80
ENTRYPOINT ["./main"]
