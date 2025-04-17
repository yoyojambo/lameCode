FROM golang:1.23

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
    # Exclusively copy source code directories
COPY ./cmd/ ./cmd
COPY ./platform/ ./platform/
COPY ./web/ ./web/
RUN go build -v -o /usr/local/bin/app ./cmd/server/

ARG DB_URL
ARG DB_TOKEN

CMD app --remote --db-url=$DB_URL --token=$DB_TOKEN
