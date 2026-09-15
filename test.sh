#!/usr/bin/env bash

# El script se detiene si cualquier comando falla
set -e

# Parametros de PostgreSQL
CONTAINER_NAME="tufutbol-postgres-test"
VOLUME_NAME="tufutbol-db-data"
DB_USER="u"
DB_PASS="p"
DB_NAME="d"
DB_PORT="5432"

cleanup() {
    echo "==> Eliminando contenedor y volumen"
    docker rm -fv "$CONTAINER_NAME" >/dev/null 2>&1 || true
    docker volume rm "$VOLUME_NAME" >/dev/null 2>&1 || true
}

# Garantiza que la funcion cleanup se ejecute siempre
trap cleanup EXIT

echo "==> Limpiando contenedores y volumenes anteriores"
cleanup

echo "==> Ejecutando sqlc generate"
sqlc generate

echo "==> Compilando el proyecto"
go build ./...

echo "==> Creando volumen y levantando PostgreSQL"
docker volume create "$VOLUME_NAME" >/dev/null
docker run --name "$CONTAINER_NAME" \
    -e POSTGRES_USER="$DB_USER" \
    -e POSTGRES_PASSWORD="$DB_PASS" \
    -e POSTGRES_DB="$DB_NAME" \
    -v "$VOLUME_NAME":/var/lib/postgresql \
    -p "$DB_PORT":5432 \
    -d postgres:alpine

echo "==> Esperando que PostgreSQL acepte conexiones"
until docker exec "$CONTAINER_NAME" pg_isready -h 127.0.0.1 -U "$DB_USER" -d "$DB_NAME" >/dev/null 2>&1; do
    sleep 0.5
done

echo "==> Aplicando esquema de base de datos"
docker exec -i "$CONTAINER_NAME" psql -h 127.0.0.1 -U "$DB_USER" -d "$DB_NAME" < db/schema/schema.sql

echo "==> Ejecutando tests"
go test -v ./...
