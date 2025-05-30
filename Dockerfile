FROM golang:1.23-alpine

COPY . /app/

ENV MONGO_URI=mongodb+srv://farmaanmalik6348:gBW3KC92dAMUtayo@cluster0.jzitdxc.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0
ENV REDIS_PW=YH8Lpp8nDLixrbkB4kJZsnRvtrbnwbGK
ENV REDIS_ADDR=redis-16663.c305.ap-south-1-1.ec2.redns.redis-cloud.com:16663

WORKDIR /app

RUN go mod tidy

RUN go build -o bin/build/gollama ./main.go

CMD ["./bin/build/gollama"]