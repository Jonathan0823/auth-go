# NOTE: If you need timezone accuracy later, add `import _ "time/tzdata"` in Go code.
FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build static binary
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -trimpath -ldflags "-s -w" -o /server .

# --- Runner (tiny) ---
FROM gcr.io/distroless/static-debian12

# Optional: make explicit (distroless defaults to nonroot)
USER nonroot:nonroot

COPY --from=builder /server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
