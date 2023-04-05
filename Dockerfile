ARG IMAGE_TYPE="production"
ARG EXTRA_PACKAGES="postgresql-client git bash vim"
ARG DEV_PACKAGES="curl netcat-openbsd iputils procps git"
ARG BUILD_PACKAGES="python3 g++ make git"

FROM node:18-alpine AS base

WORKDIR /app

FROM base AS build-base

COPY package*.json yarn.lock /app/

ARG EXTRA_PACKAGES
ARG DEV_PACKAGES
ARG BUILD_PACKAGES
ARG IMAGE_TYPE

RUN set -x \
  && apk update && apk add --no-cache $EXTRA_PACKAGES \
  && apk add --no-cache --virtual .build-deps $BUILD_PACKAGES \
  && yarn install \
  && [ "production" = "${IMAGE_TYPE}" ] && apk del .build-deps || echo

COPY . /app
RUN yarn build

FROM build-base AS base-production
RUN echo " -> Building production image" \
  && set -x \
  && apk add --no-cache --virtual .build-deps $BUILD_PACKAGES \
  && yarn install --production \
  && [ "production" = "${IMAGE_TYPE}" ] && apk del .build-deps || echo \
  && rm -rf src lib do

FROM build-base AS base-development
RUN echo " -> Building development image" \
  && set -x \
  && apk add --no-cache $DEV_PACKAGES

FROM base-${IMAGE_TYPE} AS final

ARG BUILD_HASH
ARG IMAGE_TYPE
ENV PROMPT="\[\e[35m\]buggins>\[\e[m\] "

RUN echo "Chosen build is ${IMAGE_TYPE}..." \
  && echo $BUILD_HASH >> /.buildinfo \
  && echo 'PS1=$PROMPT' >> ~/.bashrc \
  && echo "export PATH=$PATH:/app/node_modules/.bin" >> ~/.bashrc \
  && echo 'source /app/environment/local.env 2> /dev/null' >> ~/.bashrc

ENTRYPOINT [ "/app/docker/entrypoint.sh" ]
CMD [ "/app/docker/api/command.sh" ]
