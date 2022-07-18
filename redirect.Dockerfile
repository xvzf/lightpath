# syntax=docker/dockerfile:1.2
FROM alpine:3.15.4
RUN apk add -U --no-cache iptables ip6tables

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/redirect /

USER 0
ENTRYPOINT ["/redirect"]
