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
COPY data/ /app/data
COPY authorized_keys /app/authorized_keys
COPY ssh_keys   /app/ssh_keys

EXPOSE 2222
CMD ["/bin/sh", "-c", "echo 'Listing /app:' && ls -al /app && echo 'Listing /data:' && ls -al /data && echo 'Listing /data/ssh:' && ls -al /data/ssh && sleep infinity"]

CMD ["./ssh-honeypot"]