FROM golang:1.15 as builder
ENV GO111MODULE=on
COPY ./ /src/project
WORKDIR /src/project
RUN go build -o '/bin/gocrane' './'

FROM alpine:3.12
COPY --from=builder /bin/gocrane /bin/gocrane
