FROM golang:1.20 AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
ENV TZ=Asia/Taipei
WORKDIR /app
COPY --from=builder /app/main /app/main
COPY static/ static/
COPY templates/ templates/
EXPOSE $PORT
CMD ["./main"]

