# syntax=docker/dockerfile:1
FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY *.go ./
RUN go build -o /posmoni cmd/main.go


FROM scratch

WORKDIR /

COPY --from=build /posmoni /posmoni

ENTRYPOINT ["/posmoni"]