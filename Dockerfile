FROM golang:1.23.0-alpine3.20 AS build-base

WORKDIR /app

COPY . .

ENV GOPATH=/go
ENV PATH="${PATH}:/go/bin"
RUN set -x \
    && apk add --no-cache build-base=0.5-r3 make=4.4.1-r2 \
    && make release-docker

FROM gcr.io/distroless/static-debian12:9efbcaacd8eac4960b315c502adffdbf3398ce62

WORKDIR /
COPY --from=build-base /app/bin/bot /bot

CMD ["/bot", "start"]
