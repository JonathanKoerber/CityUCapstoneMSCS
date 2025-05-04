# Use the official Go image as a builder
FROM golang:1.24 AS builder

WORKDIR /app

# Copy go mod files and download dependencies first (cache-friendly)
COPY go.mod go.sum ./

RUN go mod download

# Copy the actual code
COPY . .

COPY authorized_keys /app/authorized_keys

# Build the Go binary
RUN go build -o ssh-honeypot app/main.go

# Final stage
FROM debian:bookworm

WORKDIR /app

COPY --from=builder /app/ssh-honeypot .

COPY authorized_keys /app/authorized_keys
COPY ssh_keys   /app/ssh_keys

EXPOSE 2222

CMD ["./ssh-honeypot"]