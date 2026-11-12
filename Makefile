# Comandos del proyecto TuFutbol

# La configuración (usuario y clave de la base, puertos) está en .env.
# Si no existe, make lo crea copiando .env.example (regla ".env" de abajo).
-include .env
export

# Las recetas usan bash (trap para limpiar al terminar con Ctrl+C)
SHELL := /bin/bash

APP_URL := http://localhost:$(API_PORT)
# URL de la base vista desde esta máquina (la usan make run-app y make test)
DATABASE_URL_LOCAL := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@localhost:$(DB_PORT)/$(POSTGRES_DB)?sslmode=disable

BINARIO := bin/tufutbol

.PHONY: run-app generate build test test-web up down logs clean

# Ejecuta la aplicación con todas las tareas previas y posteriores:
#   1. Tareas previas: .env, sqlc generate, templ generate, compilación, borrar
#      contenedores y volúmenes anteriores, levantar la base y esperar a que esté lista.
#   2. Ejecución del servidor en http://localhost:$(API_PORT) (se detiene con Ctrl+C).
#   3. Tareas posteriores: borrar contenedores y volúmenes (siempre, aunque algo falle).
# La limpieza ignora los Ctrl+C: el "docker compose down" corre en otra sesión
# (setsid), así un segundo Ctrl+C no lo corta a la mitad.
run-app: build
	@echo "==> Borrando contenedores y volúmenes anteriores"
	docker compose down -v
	@limpiar() { \
		trap '' INT TERM; \
		echo; echo "==> Borrando contenedores y volúmenes"; \
		if command -v setsid >/dev/null; then setsid -w docker compose down -v; else docker compose down -v; fi; \
	}; \
	trap limpiar EXIT; \
	trap 'exit 0' INT TERM; \
	echo "==> Levantando la base de datos y esperando a que esté lista"; \
	docker compose up -d --build --wait db || exit 1; \
	echo "==> TuFutbol corriendo en $(APP_URL) (Ctrl+C para terminar)"; \
	DATABASE_URL='$(DATABASE_URL_LOCAL)' PORT='$(API_PORT)' ./$(BINARIO)

# Genera el código Go: acceso a datos con sqlc (db/sqlc) y vistas con templ
# (views/*_templ.go). El código generado no se sube al repo.
# templ se ejecuta con "go run" en la misma versión que la librería de go.mod,
# así no hace falta instalarlo (la primera vez tarda un poco más porque lo descarga).
generate:
	sqlc generate
	go run github.com/a-h/templ/cmd/templ@$$(go list -m -f '{{.Version}}' github.com/a-h/templ) generate

# Compila el proyecto
build: generate
	go build -o $(BINARIO) .

# Tests de Go: capa de datos contra PostgreSQL en Docker, validación del
# formulario y componentes templ
test: generate
	docker compose down -v
	docker compose up -d --build --wait db && DATABASE_URL='$(DATABASE_URL_LOCAL)' go test -count=1 -v ./...; \
	resultado=$$?; \
	if [ $$resultado -ne 0 ]; then docker compose logs --tail 20; fi; \
	docker compose down -v; \
	exit $$resultado

# Prueba completa de la aplicación web: levanta todo en Docker, prueba cada ruta
# con curl (requests.sh) y borra todo (aunque falle). Si algo falla, antes de
# borrar muestra el log de los contenedores.
test-web: build
	docker compose down -v
	docker compose up -d --build --wait && bash requests.sh $(APP_URL); \
	resultado=$$?; \
	if [ $$resultado -ne 0 ]; then docker compose logs --tail 20; fi; \
	docker compose down -v; \
	exit $$resultado

# Levanta la base y la aplicación en Docker, en segundo plano, y espera a que estén listas
up: generate
	docker compose up -d --build --wait

# Detiene y borra los contenedores, la red y el volumen de la base
down:
	docker compose down -v

# Muestra los logs de la aplicación (cuando corre en Docker con make up)
logs:
	docker compose logs -f api

# Borra el binario y el código generado
clean:
	rm -rf bin db/sqlc views/*_templ.go

.env:
	cp .env.example .env
