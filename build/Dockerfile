# syntax=docker/dockerfile:1
ARG GO_VERSION=1.22.2-alpine

# build
FROM golang:${GO_VERSION} AS builder

WORKDIR /app

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,source=go.sum,target=go.sum \
  --mount=type=bind,source=go.mod,target=go.mod \
  go mod download

RUN --mount=type=cache,target=/go/pkg/mod/ \
  --mount=type=bind,target=. \
  GOOS=linux go build -ldflags="-s -w" -trimpath -o /bin/server ./cmd/main.go

# deployment
FROM golang:${GO_VERSION}

COPY --from=builder /bin/server /bin/server

EXPOSE 80

CMD ["/bin/server"]
