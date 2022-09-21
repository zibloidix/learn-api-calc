# Dockerfile для запуска проекта
# 1. docker image build -t learn-api-calc-img:1.0.0 .
# 2. docker container run -d --name learn-api-calc-app -e PORT=3001 -p 3001:3001 learn-api-calc-img:1.0.0
#    docker image rm learn-api-calc-img:1.0.0
#    docker container rm -f learn-api-calc-app
FROM golang:1.17.7-alpine3.15
RUN apk add build-base
WORKDIR /usr/src/app
COPY go.mod go.sum* ./
RUN go mod download
EXPOSE 3001
COPY . .
RUN go build main.go
CMD ["./main"]