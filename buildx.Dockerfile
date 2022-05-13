# syntax=docker/dockerfile:1.2

FROM alpine

RUN apk --no-cache --no-progress add ca-certificates tzdata \
    && rm -rf /var/cache/apk/*

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/controlplane /

USER 65534
ENTRYPOINT ["/controlplane"]