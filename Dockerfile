# syntax=docker/dockerfile:1

# Build on the host's native platform and cross-compile for the target, so
# multi-platform builds don't need emulation. Keep the Go version in sync
# with mise.toml.
FROM --platform=$BUILDPLATFORM golang:1.27.2-alpine AS build
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/haggar .

FROM scratch
COPY --from=build /out/haggar /go/bin/haggar
# nobody:nogroup
USER 65534:65534
ENTRYPOINT ["/go/bin/haggar"]
