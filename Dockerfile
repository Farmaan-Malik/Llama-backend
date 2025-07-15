FROM golang:1.23-alpine as build

WORKDIR /app

COPY ./go.mod ./go.sum /app/


RUN go mod download

COPY . /app/

RUN go build -o bin/build/gollama ./cmd/api

FROM alpine

COPY --from=build /app/bin/build/gollama /app/
# COPY --from=build /app/.env /app/.env

EXPOSE 8080

CMD ["/app/gollama"]