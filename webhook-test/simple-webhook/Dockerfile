FROM golang:1.22 as builder

ENV GOPROXY=https://goproxy.io
# ENV GOPROXY=https://goproxy.cn
WORKDIR /app
COPY . .

RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o webhook .

FROM alpine:3.18
COPY --from=builder /app/webhook /webhook
COPY ./certs /etc/webhook/certs

ENTRYPOINT ["/webhook"]
