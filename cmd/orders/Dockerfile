# use alpine instead of ubuntu (smaller footprint)
FROM golang:1.17-alpine AS dev
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY ./cmd/orders .
WORKDIR /app/pkg
COPY ./pkg .
WORKDIR /app
RUN go build -o main .

# I think expose should only be used in dev??
EXPOSE 8000

CMD ["./main"]

# additional steps for prod after completing dev steps (?)
#FROM dev AS prod
#WORKDIR /app