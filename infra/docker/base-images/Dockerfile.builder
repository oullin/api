ARG GO_VERSION=1.25.5
ARG GO_IMAGE_VARIANT=alpine3.23
ARG GO_IMAGE_DIGEST=sha256:ac09a5f469f307e5da71e766b0bd59c9c49ea460a528cc3e6686513d64a6f1fb

FROM golang:${GO_VERSION}-${GO_IMAGE_VARIANT}@${GO_IMAGE_DIGEST}

# Pin direct package artifact URLs so the builder base can be rebuilt even if the
# rolling Alpine package index advances past the desired revisions.
ARG TARGETARCH
ARG APK_BASE_URL=https://dl-cdn.alpinelinux.org/alpine/v3.23/main

RUN case "${TARGETARCH}" in \
        arm64) apk_arch="aarch64" ;; \
        amd64) apk_arch="x86_64" ;; \
        *) echo "Unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    wget -O /tmp/build-base.apk "${APK_BASE_URL}/${apk_arch}/build-base-0.5-r3.apk" && \
    wget -O /tmp/pkgconf.apk "${APK_BASE_URL}/${apk_arch}/pkgconf-2.5.1-r0.apk" && \
    wget -O /tmp/libwebp-dev.apk "${APK_BASE_URL}/${apk_arch}/libwebp-dev-1.6.0-r0.apk" && \
    apk add --no-cache \
        /tmp/build-base.apk \
        /tmp/pkgconf.apk \
        /tmp/libwebp-dev.apk && \
    rm -f /tmp/build-base.apk /tmp/pkgconf.apk /tmp/libwebp-dev.apk
