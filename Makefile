# Comandos del proyecto TuFutbol

# La configuración (usuario y clave de la base, puertos) está en .env.
# Si no existe, make lo crea copiando .env.example (regla ".env" de abajo).
-include .env
export

API_URL := http://localhost:$(API_PORT)
# URL de la base vista desde esta máquina (la usa make test)
DATABASE_URL_LOCAL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable

.PHONY: test-api test generate build up down logs

# Prueba completa de la API: levanta todo, prueba cada endpoint y borra todo (aunque falle).
# Si algo falla, antes de borrar muestra el log de los contenedores.
test-api: build
	docker compose down -v
	docker compose up -d --build --wait && bash requests.sh $(API_URL); \
	resultado=$$?; \
	if [ $$resultado -ne 0 ]; then docker compose logs --tail 20; fi; \
	docker compose down -v; \
	exit $$resultado

# Tests de Go de la capa de datos contra PostgreSQL en Docker
test: generate
	docker compose down -v
	docker compose up -d --build --wait db && DATABASE_URL='$(DATABASE_URL_LOCAL)' go test -count=1 -v ./...; \
	resultado=$$?; \
	if [ $$resultado -ne 0 ]; then docker compose logs --tail 20; fi; \
	docker compose down -v; \
	exit $$resultado

# Genera el código de acceso a datos con sqlc (db/sqlc no se sube al repo)
generate:
	sqlc generate

# Compila el proyecto
build: generate
	go build ./...

# Levanta la base y la API en segundo plano y espera a que estén listas
up: generate
	docker compose up -d --build --wait

# Detiene y borra los contenedores, la red y el volumen de la base
down:
	docker compose down -v

# Muestra los logs de la API
logs:
	docker compose logs -f api

.env:
	cp .env.example .env
