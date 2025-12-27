FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum first for better caching
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy the rest of the application
COPY backend/ ./
COPY . ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o quevadis-server .

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/quevadis-server .
COPY --from=builder /app/index.html .
COPY --from=builder /app/style.css .
COPY --from=builder /app/script.js .
COPY --from=builder /app/multiplayer.js .

EXPOSE 8080

CMD ["./quevadis-server"]
