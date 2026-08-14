FROM golang:1.26.1-bookworm AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /app/cmd/urlshortener/exe /app/cmd/urlshortener

FROM alpine:3.23.5
WORKDIR /app
COPY --from=build /app/cmd/urlshortener/exe /app
CMD ["/app/exe"]
