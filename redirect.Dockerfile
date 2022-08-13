# syntax=docker/dockerfile:1.2
FROM alpine:latest

RUN apk add --no-cache iptables; \
    rm /sbin/iptables; rm /sbin/iptables-save; rm /sbin/iptables-restore; \
    ln -s /sbin/iptables-nft /sbin/iptables; \
    ln -s /sbin/iptables-nft-save /sbin/iptables-save; \
    ln -s /sbin/iptables-nft-restore /sbin/iptables-restore;

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/redirect /

USER 0
ENTRYPOINT ["/redirect"]
