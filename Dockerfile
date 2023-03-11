FROM docker.io/golang:1.20-alpine AS builder

ARG CLC_VERSION="v5.2.0"
ENV CLC_VERSION=${CLC_VERSION}

WORKDIR /usr/src/app

COPY . .
RUN \
    apk update &&\
    apk add make &&\
    make

FROM docker.io/alpine:3.17

WORKDIR /

COPY --from=builder /usr/src/app/build/clc /

ENTRYPOINT ["/clc"]
CMD ["version"]

