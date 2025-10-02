# ---------- STAGE 1: Build ----------
FROM golang:1.25 AS builder
ENV CGO_ENABLED=0 GO111MODULE=on

WORKDIR /src

# copying module files first for layer caching and then copying the common files since they are a dependency
COPY services/auth/go.mod services/auth/go.sum /src/services/auth/
COPY common /src/common

WORKDIR /src/services/auth
RUN go mod download

# copying the rest of the auth service
COPY services/auth /src/services/auth

# building the auth service
WORKDIR /src/services/auth
RUN ["go", "build", "-trimpath", "-ldflags=-s -w", "-o", "/app/auth", "./cmd"]


# ---------- STAGE 2: Run ----------
FROM gcr.io/distroless/base-debian12

USER nonroot:nonroot
WORKDIR /app

COPY --from=builder /app/auth /app/auth
COPY services/auth/config /app/config

EXPOSE 8084

ENTRYPOINT ["/app/auth", "-config=/app/config/docker.yaml"]