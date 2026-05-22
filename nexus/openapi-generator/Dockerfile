# SPDX-FileCopyrightText: (C) 2025 Intel Corporation
# SPDX-License-Identifier: Apache-2.0

FROM --platform=$BUILDPLATFORM golang:1.25-bookworm AS builder

ARG TARGETARCH

COPY . /app
WORKDIR /app/openapi-generator
RUN GOPROXY=https://proxy.golang.org,direct CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -buildvcs=false -o bin/openapi-generator .

FROM photon:5.0
WORKDIR /bin
COPY --from=builder /app/openapi-generator/bin/openapi-generator .
USER 65532:65532
