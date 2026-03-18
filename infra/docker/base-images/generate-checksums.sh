#!/usr/bin/env bash
#
# Generate SHA256 checksum files for pinned APK packages used by the builder
# and runtime base images. Run this script whenever you update the package
# list in Dockerfile.builder or Dockerfile.runtime.
#
# Usage: ./generate-checksums.sh
#
set -euo pipefail

# Override with APK_BASE_URL to point at a local mirror or CDN cache.
BASE_URL="${APK_BASE_URL:-https://dl-cdn.alpinelinux.org/alpine/v3.23/main}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
CHECKSUMS_DIR="${SCRIPT_DIR}/checksums"
WORK_DIR="$(mktemp -d)"

trap 'rm -rf "${WORK_DIR}"' EXIT

mkdir -p "${CHECKSUMS_DIR}"

# Package lists — these must stay in sync with the download loops in
# Dockerfile.builder and Dockerfile.runtime. When you add, remove, or bump a
# package in a Dockerfile, mirror the change here.
BUILDER_PACKAGES=(
    binutils-2.45.1-r0.apk
    file-5.46-r2.apk
    fortify-headers-1.1-r5.apk
    g++-15.2.0-r2.apk
    gcc-15.2.0-r2.apk
    gmp-6.3.0-r4.apk
    isl26-0.26-r1.apk
    jansson-2.14.1-r0.apk
    libatomic-15.2.0-r2.apk
    libgcc-15.2.0-r2.apk
    libgomp-15.2.0-r2.apk
    libmagic-5.46-r2.apk
    libsharpyuv-1.6.0-r0.apk
    libstdc++-15.2.0-r2.apk
    libstdc++-dev-15.2.0-r2.apk
    libwebp-1.6.0-r0.apk
    libwebpdecoder-1.6.0-r0.apk
    libwebpdemux-1.6.0-r0.apk
    libwebp-dev-1.6.0-r0.apk
    libwebpmux-1.6.0-r0.apk
    make-4.4.1-r3.apk
    mpc1-1.3.1-r1.apk
    mpfr4-4.2.2-r0.apk
    musl-1.2.5-r21.apk
    musl-dev-1.2.5-r21.apk
    patch-2.8-r0.apk
    pkgconf-2.5.1-r0.apk
    zlib-1.3.2-r0.apk
    zstd-libs-1.5.7-r2.apk
)

RUNTIME_PACKAGES=(
    libsharpyuv-1.6.0-r0.apk
    libwebp-1.6.0-r0.apk
    musl-1.2.5-r21.apk
    tzdata-2026a-r0.apk
    zlib-1.3.2-r0.apk
)

# Architectures for which we generate checksums (multi-platform support).
ARCHES=(aarch64 x86_64)

generate_checksums() {
    local image_type="$1"
    shift
    local packages=("$@")

    for arch in "${ARCHES[@]}"; do
        local outfile="${CHECKSUMS_DIR}/${image_type}-${arch}.sha256"
        local dl_dir="${WORK_DIR}/${image_type}/${arch}"
        mkdir -p "${dl_dir}"

        printf "Downloading %s packages for %s...\n" "${image_type}" "${arch}"

        for pkg in "${packages[@]}"; do
            local url="${BASE_URL}/${arch}/${pkg}"
            if ! curl -fsSL -o "${dl_dir}/${pkg}" "${url}"; then
                printf "ERROR: failed to download %s\n" "${url}" >&2
                exit 1
            fi
        done

        printf "Computing checksums for %s/%s...\n" "${image_type}" "${arch}"
        (cd "${dl_dir}" && sha256sum -- *.apk | sort -k2) > "${outfile}"

        printf "  -> %s\n" "${outfile}"
    done
}

generate_checksums "builder" "${BUILDER_PACKAGES[@]}"
generate_checksums "runtime" "${RUNTIME_PACKAGES[@]}"

printf "\nDone. Checksum files written to %s/\n" "${CHECKSUMS_DIR}"
