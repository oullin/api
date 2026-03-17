# Reproducible API Base Images

These Dockerfiles hold the pinned Alpine package set that the application image
depends on.

The canonical version inputs for these images live in `infra/makefile/build.mk`.
`make ensure-base-images` and `make build-base-images` pass the Go version,
Alpine version, base-image digests, and APK repository URL into the Dockerfiles
explicitly so the image tags and build inputs stay aligned.

## Images

- `Dockerfile.builder`: Go builder base with `build-base`, `pkgconf`, and `libwebp-dev`
- `Dockerfile.runtime`: runtime base with `tzdata` and `libwebp`

## Why direct `.apk` downloads

`apk add package=version` breaks once Alpine advances the package index past a
specific revision. These base Dockerfiles download the exact `.apk` artifacts by
filename, including the transitive dependency closure needed by the builder and
runtime images, and install them with `apk --no-network` so rebuilds do not
consult the live package index.

## Checksum verification

Every downloaded `.apk` is verified against SHA256 checksums stored in `checksums/`.
The local APKINDEX is signed with an ephemeral RSA key so `apk add` runs without
`--allow-untrusted`.

## Refresh workflow

1. Update the package filenames in these Dockerfiles when you intentionally want
   a new pinned package set.
2. Regenerate checksums: `make generate-apk-checksums` (or run
   `./infra/docker/base-images/generate-checksums.sh` directly).
3. Run `make build-base-images`.
4. Run `make push-base-images`.
5. Copy the printed `builder=` and `runtime=` digests into the `BUILDER_BASE_IMAGE`
   and `RUNTIME_BASE_IMAGE` build args when you are ready to consume registry-pinned
   references.
