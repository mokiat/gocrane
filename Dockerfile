FROM golang:1.18 as builder
COPY ./ /src/project
WORKDIR /src/project
RUN CGO_ENABLED=0 go build -o '/bin/gocrane' './'

FROM alpine:3.15
COPY --from=builder /bin/gocrane /bin/gocrane
