FROM golang:1.26 AS builder
COPY ./ /src/project
WORKDIR /src/project
RUN CGO_ENABLED=0 go build -o '/bin/gocrane' './'

FROM alpine:3.23
COPY --from=builder /bin/gocrane /bin/gocrane
