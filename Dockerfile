FROM golang:1.17.11-stretch AS builder

WORKDIR /src
COPY . .
RUN go build -o app main.go

FROM ubuntu:20.04 AS prod
COPY --from=builder /src/app /app
EXPOSE 9000

ENTRYPOINT ["/app"]
