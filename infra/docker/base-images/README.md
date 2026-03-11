# Reproducible API Base Images

These Dockerfiles hold the pinned Alpine package set that the application image
depends on.

## Images

- `Dockerfile.builder`: Go builder base with `build-base`, `pkgconf`, and `libwebp-dev`
- `Dockerfile.runtime`: runtime base with `tzdata` and `libwebp`

## Why direct `.apk` downloads

`apk add package=version` breaks once Alpine advances the package index past a
specific revision. These base Dockerfiles download the exact `.apk` artifacts by
filename so the pinned package set can still be rebuilt.

## Refresh workflow

1. Update the package filenames in these Dockerfiles when you intentionally want
   a new pinned package set.
2. Run `make build-base-images`.
3. Run `make push-base-images`.
4. Copy the printed `builder=` and `runtime=` digests into the `BUILDER_BASE_IMAGE`
   and `RUNTIME_BASE_IMAGE` build args when you are ready to consume registry-pinned
   references.
