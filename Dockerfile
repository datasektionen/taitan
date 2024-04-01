FROM golang:1.22-alpine3.19 AS build

WORKDIR /app/

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY taitan.go ./
COPY pages ./pages
COPY fuzz ./fuzz
COPY anchor ./anchor

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/taitan taitan.go

FROM alpine:3.19

WORKDIR /app/

RUN apk add --no-cache git
COPY --from=build /app/taitan ./

CMD ["./taitan", "-v"]
