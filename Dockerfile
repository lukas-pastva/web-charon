FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /charon ./cmd/charon

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

COPY --from=builder /charon /charon

EXPOSE 8080

ENTRYPOINT ["/charon"]
