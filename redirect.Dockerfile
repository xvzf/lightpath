# syntax=docker/dockerfile:1.2

FROM ubuntu:jammy

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && \
  apt-get install --no-install-recommends -y \
  iptables \
  && apt-get upgrade -y \
  && apt-get clean \
  && rm -rf  /var/log/*log /var/lib/apt/lists/* /var/log/apt/* /var/lib/dpkg/*-old /var/cache/debconf/*-old \
  && update-alternatives --set iptables /usr/sbin/iptables-nft \
  && update-alternatives --set ip6tables /usr/sbin/ip6tables-nft

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/redirect /

USER 0
ENTRYPOINT ["/redirect"]
