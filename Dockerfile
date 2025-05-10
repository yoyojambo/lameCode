FROM golang:1.24

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download
    # Exclusively copy source code directories
COPY ./cmd/ ./cmd
COPY ./platform/ ./platform/
COPY ./web/ ./web/
RUN go build -v -o /usr/local/bin/app ./cmd/server/

ARG DB_URL
ENV DB_URL=$DB_URL
ARG DB_TOKEN
ENV DB_TOKEN=$DB_TOKEN

CMD app --remote --install-wasmer --db-url=$DB_URL --token=$DB_TOKEN
