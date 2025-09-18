FROM debian:12-slim AS kubectl-downloader
RUN apt-get update && apt-get install -y curl ca-certificates \
 && curl -LO "https://dl.k8s.io/release/v1.31.1/bin/linux/amd64/kubectl" \
 && chmod +x kubectl

FROM gcr.io/distroless/static-debian12:nonroot
USER nobody

# Now copy the static shell into base image.
COPY --from=busybox:1.37.0-uclibc /bin/sh /bin/sh

# Copy kubectl from the downloader stage
COPY --from=kubectl-downloader /kubectl /usr/local/bin/kubectl

# Set the working directory in the container
WORKDIR /app

COPY datamodel_installer.sh /datamodel_installer.sh
COPY build/crds /crds
COPY build/clusterroles /clusterroles

COPY IMAGE /IMAGE
COPY NAME /NAME

CMD ["/bin/sh","/datamodel_installer.sh"]
