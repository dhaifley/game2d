FROM golang:latest AS build-stage

ARG VERSION="0.1.1"

ARG MAIN_PACKAGE="./cmd/game2d-api"

WORKDIR /go/src/github.com/dhaifley/game2d

ADD . .

RUN go mod download

RUN CGO_ENABLED=1 go build -v -o /go/bin/game2d \
    -ldflags="-X github.com/dhaifley/game2d/server.Version=$VERSION" \
  $MAIN_PACKAGE

FROM alpine:latest AS certs-stage

RUN apk --update add ca-certificates

FROM scratch AS release-stage

ARG PORT=8080

WORKDIR /

COPY --from=certs-stage /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

COPY --from=build-stage /go/bin/game2d /

COPY --from=build-stage /go/src/github.com/dhaifley/game2d/certs/* /certs/

EXPOSE $PORT/tcp

ENTRYPOINT ["/game2d"]
