FROM mokiat/gocrane:latest AS gocrane

FROM golang:1.16
COPY --from=gocrane /bin/gocrane /bin/gocrane
COPY . /src/project
WORKDIR /src/project

ENV GOCRANE_SOURCES='./'
ENV GOCRANE_MAIN='./cmd/example'
ENV GOCRANE_BINARY='/bin/example'
ENV GOCRANE_DIGEST='/bin/example.sha256'
RUN gocrane build

CMD ["gocrane", "run"]
