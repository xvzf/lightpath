# syntax=docker/dockerfile:1.2
FROM ubuntu:xenial

RUN apk add --no-cache iptables && rm /sbin/iptables && ln -s /sbin/iptables /sbin/iptables-nft

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/redirect /

USER 0
ENTRYPOINT ["/redirect"]
