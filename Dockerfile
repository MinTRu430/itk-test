# build stage
FROM golang:1.24-alpine AS build
WORKDIR /app
ENV CGO_ENABLED=0
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /wallet-app ./cmd/itk

# final
FROM alpine:3.18
RUN apk add --no-cache ca-certificates
COPY --from=build /wallet-app /wallet-app
COPY config.env /config.env
ENV GIN_MODE=release
EXPOSE 8080
CMD ["/wallet-app"]
