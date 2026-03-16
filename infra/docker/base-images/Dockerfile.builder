ARG GO_VERSION
ARG GO_IMAGE_VARIANT
ARG GO_IMAGE_DIGEST

FROM golang:${GO_VERSION}-${GO_IMAGE_VARIANT}@${GO_IMAGE_DIGEST}

# Pin direct package artifact URLs so the builder base can be rebuilt even if the
# rolling Alpine package index advances past the desired revisions.
# TARGETARCH is auto-set by BuildKit. Fall back to the running Alpine arch when
# building without BuildKit-provided platform args.
ARG TARGETARCH
ARG APK_BASE_URL

RUN target_arch="${TARGETARCH}"; \
    if [ -z "${target_arch}" ]; then \
        case "$(apk --print-arch)" in \
            aarch64) target_arch="arm64" ;; \
            x86_64) target_arch="amd64" ;; \
            *) echo "Unsupported Alpine arch: $(apk --print-arch)" >&2; exit 1 ;; \
        esac; \
    fi && \
    case "${target_arch}" in \
        arm64) apk_arch="aarch64" ;; \
        amd64) apk_arch="x86_64" ;; \
        *) echo "Unsupported TARGETARCH: ${TARGETARCH}" >&2; exit 1 ;; \
    esac && mkdir -p "/tmp/local-repo/${apk_arch}" && \
    for apk_pkg in \
        binutils-2.45.1-r0.apk \
        file-5.46-r2.apk \
        fortify-headers-1.1-r5.apk \
        g++-15.2.0-r2.apk \
        gcc-15.2.0-r2.apk \
        gmp-6.3.0-r4.apk \
        isl26-0.26-r1.apk \
        jansson-2.14.1-r0.apk \
        libatomic-15.2.0-r2.apk \
        libgcc-15.2.0-r2.apk \
        libgomp-15.2.0-r2.apk \
        libmagic-5.46-r2.apk \
        libsharpyuv-1.6.0-r0.apk \
        libstdc++-15.2.0-r2.apk \
        libstdc++-dev-15.2.0-r2.apk \
        libwebp-1.6.0-r0.apk \
        libwebpdecoder-1.6.0-r0.apk \
        libwebpdemux-1.6.0-r0.apk \
        libwebp-dev-1.6.0-r0.apk \
        libwebpmux-1.6.0-r0.apk \
        make-4.4.1-r3.apk \
        mpc1-1.3.1-r1.apk \
        mpfr4-4.2.2-r0.apk \
        musl-1.2.5-r21.apk \
        musl-dev-1.2.5-r21.apk \
        patch-2.8-r0.apk \
        pkgconf-2.5.1-r0.apk \
        zlib-1.3.2-r0.apk \
        zstd-libs-1.5.7-r2.apk \
    ; do \
        wget -qO "/tmp/local-repo/${apk_arch}/${apk_pkg}" "${APK_BASE_URL}/${apk_arch}/${apk_pkg}" || exit 1; \
    done && \
    apk index --allow-untrusted -q -o "/tmp/local-repo/${apk_arch}/APKINDEX.tar.gz" /tmp/local-repo/${apk_arch}/*.apk && \
    apk add --no-cache --no-network --allow-untrusted --repositories-file /dev/null --repository /tmp/local-repo binutils file g++ gcc make musl-dev patch pkgconf libwebp-dev && \
    apk add --no-cache --no-network --allow-untrusted "/tmp/local-repo/${apk_arch}/fortify-headers-1.1-r5.apk" && \
    rm -rf /tmp/local-repo
