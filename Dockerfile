# ---- build stage ----
FROM --platform=$BUILDPLATFORM golang:1.24.1 AS builder

ARG TARGETOS=linux
ARG TARGETARCH=amd64
ARG BUILD_TAG=dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# собираем бинарь и ставим исполняемые права здесь
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w -X 'main.buildTag=${BUILD_TAG}'" \
    -o /app/bin/mpstats-sync ./cmd/server \
 && chmod 0755 /app/bin/mpstats-sync

# ---- runtime stage ----
FROM gcr.io/distroless/base-debian12
WORKDIR /app

# сразу назначаем владельца nonroot
COPY --from=builder --chown=nonroot:nonroot /app/bin/mpstats-sync /app/mpstats-sync

EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/mpstats-sync"]
