FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build

WORKDIR /src
ARG TARGETOS TARGETARCH
RUN --mount=target=. \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/moneybags /src/services/moneybags

FROM alpine

# Copy binary
COPY --from=build /out/moneybags /bin

EXPOSE 8080

CMD ["moneybags", "--dry-run"]