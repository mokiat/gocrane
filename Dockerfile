FROM golang:1.22 as builder
COPY ./ /src/project
WORKDIR /src/project
RUN CGO_ENABLED=0 go build -o '/bin/gocrane' './'

FROM alpine:3.17
COPY --from=builder /bin/gocrane /bin/gocrane
