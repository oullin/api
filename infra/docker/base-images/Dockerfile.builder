ARG GO_VERSION=1.26.1
ARG GO_IMAGE_VARIANT=alpine3.23
ARG GO_IMAGE_DIGEST=sha256:2389ebfa5b7f43eeafbd6be0c3700cc46690ef842ad962f6c5bd6be49ed82039

FROM golang:${GO_VERSION}-${GO_IMAGE_VARIANT}@${GO_IMAGE_DIGEST}

# Pin direct package artifact URLs so the builder base can be rebuilt even if the
# rolling Alpine package index advances past the desired revisions.
# TARGETARCH is auto-set by BuildKit; default to amd64 for legacy builders.
ARG TARGETARCH=amd64
ARG APK_BASE_URL=https://dl-cdn.alpinelinux.org/alpine/v3.23/main

RUN case "${TARGETARCH}" in \
        arm64) apk_arch="aarch64" ;; \
        amd64) apk_arch="x86_64" ;; \
        *) echo "Unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && \
    wget -qO /tmp/build-base.apk "${APK_BASE_URL}/${apk_arch}/build-base-0.5-r3.apk" && \
    wget -qO /tmp/pkgconf.apk "${APK_BASE_URL}/${apk_arch}/pkgconf-2.5.1-r0.apk" && \
    wget -qO /tmp/libwebp-dev.apk "${APK_BASE_URL}/${apk_arch}/libwebp-dev-1.6.0-r0.apk" && \
    apk add --no-cache \
        /tmp/build-base.apk \
        /tmp/pkgconf.apk \
        /tmp/libwebp-dev.apk && \
    rm -f /tmp/build-base.apk /tmp/pkgconf.apk /tmp/libwebp-dev.apk
