FROM golang:1.23.2-alpine3.20 AS build-base

WORKDIR /app

COPY . .

ENV GOPATH=/go
ENV PATH="${PATH}:/go/bin"
ENV CGO_ENABLED=1
RUN set -x \
    && apk add --no-cache build-base=0.5-r3 \
    && go build \
        -a \
        -tags release \
        -ldflags '-s -w -linkmode external -extldflags "-static"' \
        -o bin/bot .

FROM gcr.io/distroless/static-debian12:9efbcaacd8eac4960b315c502adffdbf3398ce62

WORKDIR /
COPY --from=build-base /app/bin/bot /bot

CMD ["/bot", "start"]
