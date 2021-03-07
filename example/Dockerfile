FROM mokiat/gocrane:latest AS gocrane

FROM golang:1.16
COPY --from=gocrane /bin/gocrane /bin/gocrane
COPY . /src/project
WORKDIR /src/project

ENV GOCRANE_RESOURCES='./public'
ENV GOCRANE_MAIN='./cmd/example'
ENV GOCRANE_BINARY='/bin/example'
RUN gocrane build -v

CMD ["gocrane", "run", "-v"]
