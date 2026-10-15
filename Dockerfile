# Etapa 1: compilar la API con la imagen oficial de Go
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /tufutbol .

# Etapa 2: imagen final, solo con el binario y la página estática
FROM alpine:3.22
WORKDIR /app
COPY --from=build /tufutbol .
COPY static ./static
EXPOSE 8080
CMD ["./tufutbol"]
