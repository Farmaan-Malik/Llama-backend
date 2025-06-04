FROM golang:1.23-alpine

COPY . /app/

WORKDIR /app

RUN go mod tidy

RUN go build -o bin/build/gollama ./main.go

CMD ["./bin/build/gollama"]