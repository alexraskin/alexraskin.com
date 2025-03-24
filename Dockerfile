FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o alexraskin-web .

FROM alpine:3.18

WORKDIR /app

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/alexraskin-web .

COPY --from=builder /app/static ./static
COPY --from=builder /app/public ./public
COPY --from=builder /app/static/robots.txt ./static/robots.txt
COPY --from=builder /app/static/sitemap.xml ./static/sitemap.xml

EXPOSE 8000

CMD ["./alexraskin-web", "-port", "8000"] 