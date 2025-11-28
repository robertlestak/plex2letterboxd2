FROM golang:1.25.4-alpine AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /bin/p2l2 cmd/p2l2/p2l2.go

FROM debian:stable
WORKDIR /app
RUN apt-get update && apt-get install -y ca-certificates nodejs npm && rm -rf /var/lib/apt/lists/*
COPY --from=builder /bin/p2l2 /bin/p2l2
RUN /bin/p2l2 -init-playwright-only && npx playwright install-deps
ENTRYPOINT ["/bin/p2l2"]
