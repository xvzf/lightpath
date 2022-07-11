# syntax=docker/dockerfile:1.2

FROM gcr.io/distroless/static

ARG TARGETPLATFORM
COPY ./dist/$TARGETPLATFORM/controlplane /

USER 65534
ENTRYPOINT ["/controlplane"]