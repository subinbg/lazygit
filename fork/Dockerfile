# syntax=docker/dockerfile:1
# The fork's toolchain, run from the repository root (see fork/dev):
#   base     the Go toolchain on Alpine
#   test     unit tests; TEST_ARGS="integration [pattern]" runs the integration tests instead,
#            those whose name (e.g. main_view/stage_diff_lines) matches the pattern if one is given
#   build    the binary for TARGET (linux-x64 | linux-arm64 | darwin-arm64)
#   export   that binary alone, for `docker build --output`
ARG GO_VERSION=1.26

FROM golang:${GO_VERSION}-alpine AS base

# coreutils: some integration tests use GNU options that BusyBox lacks (seq -f)
RUN apk add --no-cache bash coreutils git
# GOTOOLCHAIN=local fails when go.mod needs a newer Go than GO_VERSION.
ENV CGO_ENABLED=0 GOTOOLCHAIN=local
WORKDIR /work

FROM base AS source
COPY . ./

FROM source AS test
ARG TEST_ARGS=""
# The integration tests commit in their sandbox repos.
RUN git config --global user.name test && git config --global user.email test@example.com
RUN --mount=type=cache,id=lazygit-go-build,target=/root/.cache/go-build \
    set -- $TEST_ARGS && case "${1:-unit}" in \
      unit) go test ./... -short ;; \
      integration) go test -count=1 -timeout 40m pkg/integration/clients/*.go -run "TestIntegration/${2:-}" ;; \
      *) echo "TEST_ARGS: unit | integration [pattern]" >&2; exit 1 ;; \
    esac

FROM source AS build
ARG TARGET=linux-x64
ARG VERSION=dev
ARG COMMIT=unknown
RUN --mount=type=cache,id=lazygit-go-build,target=/root/.cache/go-build \
    case "$TARGET" in \
      linux-x64) export GOOS=linux GOARCH=amd64 ;; \
      linux-arm64) export GOOS=linux GOARCH=arm64 ;; \
      darwin-arm64) export GOOS=darwin GOARCH=arm64 ;; \
      *) echo "unknown TARGET '$TARGET'" >&2; exit 1 ;; \
    esac \
 && ldflags="-s -w -X main.version=$VERSION -X main.commit=$COMMIT" \
 && ldflags="$ldflags -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
 && ldflags="$ldflags -X main.buildSource=binaryRelease" \
 && mkdir -p /out && go build -trimpath -buildvcs=false -ldflags "$ldflags" -o "/out/lazygit-$TARGET" .

FROM scratch AS export
COPY --from=build /out/ /
