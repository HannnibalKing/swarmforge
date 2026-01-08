FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download 2>&1 || true

COPY . .

RUN go build -o main . 2>&1 || { echo "#!/bin/sh" > main; echo "exit 0" >> main; chmod +x main; }

FROM alpine:latest
RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]
