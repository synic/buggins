FROM golang:1.23.0-bookworm as build-base

WORKDIR /app

COPY . .

ENV GOPATH=/go
ENV PATH="${PATH}:/go/bin"
RUN mkdir /app/data
RUN set -x \
  && apt-get update \
  && apt-get -y --no-install-recommends install build-essential=12.9 \
  && make install-builddeps \
  && make release

FROM gcr.io/distroless/static-debian12:9efbcaacd8eac4960b315c502adffdbf3398ce62

WORKDIR /
COPY --from=build-base /go/bin/migrate /migrate
COPY --from=build-base /app/bin/bot /bot
COPY --from=build-base /app/internal/db/migrations /migrations
COPY --from=build-base /app/docker/start.sh /start.sh
COPY --from=build-base /app/data /data
COPY --from=busybox:1.35.0-uclibc /bin/sh /bin/sh

CMD ["sh", "/start.sh"]
