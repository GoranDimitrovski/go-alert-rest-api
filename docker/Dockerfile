FROM golang:1.22 as builder
WORKDIR /app
COPY src/go.mod .
COPY src/go.sum .
RUN go mod download
COPY /src .
RUN go build -o app

FROM ubuntu:22.04
RUN mkdir /app
WORKDIR /app
COPY --from=builder /app/app .
ENTRYPOINT ["./app"]