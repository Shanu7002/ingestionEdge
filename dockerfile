# # Stage 1: Build the static binary
# FROM golang:1.25-alpine AS builder
#
# WORKDIR /app
#
# # Cache dependency downloads
# COPY go.mod go.sum ./
# RUN go mod download
#
# # Copy source code
# COPY . .
#
# # Compile static binary with zero CGO dependencies
# RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/edge-api ./cmd/edge-api
#
# # Stage 2: Runtime Environment
# FROM alpine:3.20
#
# RUN apk add --no-cache ca-certificates
#
# # Copy the compiled binary from the builder stage
# COPY --from=builder /bin/edge-api /bin/edge-api
#
# EXPOSE 8125/udp 8080/tcp 9090/tcp
#
# ENTRYPOINT ["/bin/edge-api"]