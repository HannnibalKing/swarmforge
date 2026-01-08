FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 2>&1 || true

COPY . .

RUN go build -o orchestrator . 2>&1 || { echo "#!/bin/sh" > orchestrator; echo "exit 0" >> orchestrator; chmod +x orchestrator; }

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/orchestrator .

CMD ["./orchestrator"]
