#Build stage
FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache git
COPY go.mod go.sum ./
COPY main.go .
RUN go mod tidy
RUN go mod download

# Build the Go binary (static binary)
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o backend-app

# Final stage
FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/backend-app .
EXPOSE 8080
CMD ["./backend-app"]
