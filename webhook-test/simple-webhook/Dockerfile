FROM golang:1.20 as builder

WORKDIR /app
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook .

FROM alpine:3.18
COPY --from=builder /app/webhook /webhook
COPY ./certs /etc/webhook/certs

ENTRYPOINT ["/webhook"]
