# Etapa 1: compilar el servidor con la imagen oficial de Go.
# El código generado por sqlc y templ se genera antes con make (make up / make test-web).
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /tufutbol .

# Etapa 2: imagen final, solo con el binario y los archivos estáticos (CSS e imágenes)
FROM alpine:3.22
WORKDIR /app
COPY --from=build /tufutbol .
COPY static ./static
EXPOSE 8080
CMD ["./tufutbol"]
