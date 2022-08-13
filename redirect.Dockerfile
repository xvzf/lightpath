# syntax=docker/dockerfile:1.2
FROM xenial

RUN apt-get update && apt-get install -y \
  iptables \
  && rm -rf /var/lib/apt/lists/*


ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/redirect /

USER 0
ENTRYPOINT ["/redirect"]
