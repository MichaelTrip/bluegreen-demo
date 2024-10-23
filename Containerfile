FROM docker.io/golang:alpine3.17 AS builder

WORKDIR /src
COPY . .
RUN go build -o dist/

FROM alpine:3.17
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /src/dist/bluegreen-demo .
RUN chmod +x /app/bluegreen-demo

CMD ["/app/bluegreen-demo"]
