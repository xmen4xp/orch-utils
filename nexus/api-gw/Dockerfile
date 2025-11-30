# SPDX-FileCopyrightText: 2025 Intel Corporation
#
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.24.6-bullseye AS builder

ARG API_GW_COMPONENT_NAME
ARG HTTP_PROXY
ARG HTTPS_PROXY
ARG FTP_PROXY
ARG SOCKS_PROXY
ARG NO_PROXY
ARG USER_HOME

COPY . /app

ENV http_proxy=${HTTP_PROXY}
ENV https_proxy=${HTTPS_PROXY}
ENV ftp_proxy=${FTP_PROXY}
ENV socks_proxy=${SOCKS_PROXY}
ENV no_proxy=${NO_PROXY}

WORKDIR /app
RUN set -x && \
    apt-get update -y && \
    apt-get install -y --no-install-recommends netcat-openbsd=1.217-3 && \
    make build 

#FROM gcr.io/distroless/static-debian12:nonroot
FROM golang:1.24
# Run as non-privileged user
USER nobody

WORKDIR /bin
COPY --from=builder /app/build/bin/api-gw /usr/local/bin/api-gw

ENTRYPOINT ["api-gw"]
