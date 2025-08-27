FROM golang:1.24-alpine AS builder
ADD ./main.go /app/main.go
ADD ./go.mod /app/go.mod
ADD ./public/index.html /app/public/index.html
WORKDIR /app
RUN go build -o main .

FROM scratch
COPY --from=builder /app/main /main
ENTRYPOINT ["/main"]
