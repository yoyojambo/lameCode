FROM golang:1.24-alpine AS builder

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
    # Exclusively copy source code directories
COPY ./cmd/ ./cmd
COPY ./platform/ ./platform/
COPY ./web/ ./web/
RUN go build -o /usr/local/bin/app ./cmd/server/

FROM alpine:edge

RUN apk add --no-cache rust-wasm go
RUN apk add --no-cache -X https://dl-cdn.alpinelinux.org/alpine/edge/testing iwasm

COPY --from=builder /usr/local/bin/app /usr/local/bin/app
COPY --from=builder /usr/src/app/web/assets ./web/assets/
COPY --from=builder /usr/src/app/web/templates ./web/templates/

ARG DB_URL
ENV DB_URL=$DB_URL
ARG DB_TOKEN
ENV DB_TOKEN=$DB_TOKEN

CMD app --remote --db-url=$DB_URL --token=$DB_TOKEN
