FROM golang:1.22.0-alpine3.19 as build-base

WORKDIR /app

COPY . .

ENV GOPATH=/go
ENV PATH="${PATH}:/go/bin"
RUN set -x \
    && apk add --update nodejs npm make \
    && make install-builddeps \
    && make release \
    && rm -rf node_modules

FROM gcr.io/distroless/static-debian11

WORKDIR /
COPY --from=build-base /app/bin/blog-serve-release /blog-serve
COPY --from=build-base /app/assets /assets
COPY --from=build-base /app/articles /articles

CMD ["./blog-serve"]
