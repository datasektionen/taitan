FROM golang:1.21-alpine

WORKDIR /app/

RUN apk add --no-cache git
COPY go.mod go.sum /app/
RUN go mod download && go mod verify

COPY taitan.go /app/
COPY pages /app/pages
COPY fuzz /app/fuzz
COPY anchor /app/anchor

RUN go build taitan.go

CMD ["./taitan", "-v"]