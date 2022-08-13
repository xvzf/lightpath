# syntax=docker/dockerfile:1.2
FROM alpine:latest

RUN apk add --no-cache iptables ip6tables; \
    rm /sbin/iptables; rm /sbin/iptables-save; rm /sbin/iptables-restore; \
    ln -s /sbin/iptables-nft /sbin/iptables; \
    ln -s /sbin/iptables-nft-save /sbin/iptables-save; \
    ln -s /sbin/iptables-nft-restore /sbin/iptables-restore; \
    rm /sbin/ip6tables; rm /sbin/ip6tables-save; rm /sbin/ip6tables-restore; \
    ln -s /sbin/ip6tables-nft /sbin/ip6tables; \
    ln -s /sbin/ip6tables-nft-save /sbin/ip6tables-save; \
    ln -s /sbin/ip6tables-nft-restore /sbin/ip6tables-restore;

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/redirect /

USER 0
ENTRYPOINT ["/redirect"]
