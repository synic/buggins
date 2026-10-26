FROM golang:1.27.3-alpine AS build-base

WORKDIR /app

COPY . .

RUN go tool github.com/magefile/mage build:release

FROM gcr.io/distroless/static-debian13

WORKDIR /
COPY --from=build-base /app/bin/bot-release /bot

CMD ["/bot", "start"]
