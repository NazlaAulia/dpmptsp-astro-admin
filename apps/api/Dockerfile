# No `# syntax=` pin: it makes BuildKit fetch an external frontend on every
# build, so a registry hiccup breaks local builds. The built-in frontend
# already supports the cache mounts used below.

# Go is not installed on the host, and does not need to be: the toolchain lives
# here. `make build` and `make test` for the API work through this image.
FROM golang:1.25-alpine AS build
WORKDIR /src

# Dependencies first, so editing a .go file does not invalidate the module
# download layer.
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download

COPY . .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/api ./cmd/api

# Create the storage root here so it exists in the image owned by the runtime
# user. Docker seeds a fresh named volume from the image path, ownership
# included; without this the volume is root-owned and the nonroot process
# cannot write to it. distroless has no shell, so it cannot be done at runtime.
RUN mkdir -p /out/storage-root && chown -R 65532:65532 /out/storage-root

# distroless static: no shell, no package manager, no libc — nothing for an
# attacker to pivot to if they get code execution.
FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/api /api
COPY --from=build --chown=nonroot:nonroot /out/storage-root /var/lib/dpmptsp
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/api"]
