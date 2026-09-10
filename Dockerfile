FROM golang:1.27-alpine AS builder

RUN apk add --no-cache git

WORKDIR /build

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s" \
    -o app ./cmd/api


FROM golang:1.27-alpine AS migrator

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 go install \
    github.com/pressly/goose/v3/cmd/goose@latest

COPY migrations /migrations


FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

COPY --from=builder /build/app /app/app
COPY --from=busybox:stable-uclibc /bin/wget /usr/local/bin/wget
COPY migrations /app/migrations

ENTRYPOINT ["/app/app"]