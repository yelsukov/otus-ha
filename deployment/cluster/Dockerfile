FROM alpine:3.7

RUN apk add --update mysql-client && rm -rf /var/cache/apk/*
RUN apk add --no-cache bash

COPY setupCluster.sh setupCluster.sh