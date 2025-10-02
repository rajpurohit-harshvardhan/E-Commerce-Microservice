# ---------- STAGE 1: Build ----------
FROM golang:1.25 AS builder
ENV CGO_ENABLED=0 GO111MODULE=on

WORKDIR /src

# copying module files first for layer caching and then copying the common files since they are a dependency
COPY services/product/go.mod services/product/go.sum /src/services/product/
COPY common /src/common

WORKDIR /src/services/product
RUN go mod download

# copying the rest of the product service
COPY services/product /src/services/product

# building the product service
WORKDIR /src/services/product
RUN ["go", "build", "-trimpath", "-ldflags=-s -w", "-o", "/app/product", "./cmd"]


# ---------- STAGE 2: Run ----------
FROM gcr.io/distroless/base-debian12

USER nonroot:nonroot
WORKDIR /app

COPY --from=builder /app/product /app/product
COPY services/product/config /app/config

EXPOSE 8084

ENTRYPOINT ["/app/product", "-config=/app/config/docker.yaml"]