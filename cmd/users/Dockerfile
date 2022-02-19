# use alpine instead of ubuntu (smaller footprint)
FROM golang:1.17-alpine AS dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/users .
WORKDIR /app/pkg
COPY ./pkg .
WORKDIR /app
RUN go build -o main .

EXPOSE 8001

CMD ["./main"]