FROM golang:1.21-alpine

WORKDIR /app
COPY go.mod .
COPY main.go .

RUN go build -o node

EXPOSE 8000

ENTRYPOINT ["./node"]