FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY . .
RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux go build -o asfpc .

FROM scratch
WORKDIR /root/

COPY --from=builder /app/asfpc .
CMD ["./asfpc"]