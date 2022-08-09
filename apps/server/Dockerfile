FROM --platform=$BUILDPLATFORM golang:1.19-alpine AS build

WORKDIR /src
ARG TARGETOS TARGETARCH
RUN --mount=target=. \
  --mount=type=cache,target=/root/.cache/go-build \
  --mount=type=cache,target=/go/pkg \
  GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o /out/boompow-server /src/apps/server

FROM alpine

# Copy email templates
ADD ./apps/server/src/email/templates /src/apps/server/src/email/templates

# Copy binary
COPY --from=build /out/boompow-server /bin

EXPOSE 8080

CMD ["boompow-server", "server"]