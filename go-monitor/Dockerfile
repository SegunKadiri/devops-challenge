FROM golang:1.21-alpine AS build
WORKDIR /src
COPY go.mod .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/monitor main.go

FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /app/monitor /usr/local/bin/monitor
USER 1000:1000
ENTRYPOINT ["/usr/local/bin/monitor"]
