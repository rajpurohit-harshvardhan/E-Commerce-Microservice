# ---------- STAGE 1: Build ----------
FROM golang:1.25 AS builder
ENV CGO_ENABLED=0 GO111MODULE=on

WORKDIR /src

# copying module files first for layer caching and then copying the common files since they are a dependency
COPY services/order/go.mod services/order/go.sum /src/services/order/
COPY common /src/common

WORKDIR /src/services/order
RUN go mod download

# copying the rest of the order service
COPY services/order /src/services/order

# building the order service
WORKDIR /src/services/order
RUN ["go", "build", "-trimpath", "-ldflags=-s -w", "-o", "/app/order", "./cmd"]


# ---------- STAGE 2: Run ----------
FROM gcr.io/distroless/base-debian12

USER nonroot:nonroot
WORKDIR /app

COPY --from=builder /app/order /app/order
COPY services/order/config /app/config

EXPOSE 8083

ENTRYPOINT ["/app/order", "-config=/app/config/docker.yaml"]