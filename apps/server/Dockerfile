FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build
WORKDIR /src
ARG TARGETOS TARGETARCH
RUN --mount=target=. \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/boompow-server /src/apps/server

FROM alpine
COPY --from=build /out/boompow-server /bin

EXPOSE 8080

CMD ["boompow-server", "server"]