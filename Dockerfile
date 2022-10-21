# syntax=docker/dockerfile:1
FROM golang:1.19-bullseye AS build
WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=1 go build -o /posmoni cmd/main.go

FROM debian:stable-20221004-slim
WORKDIR /

COPY --from=build /posmoni /posmoni
ENTRYPOINT ["/posmoni"]
