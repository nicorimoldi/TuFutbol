<img src="docs/logo.svg" alt="TuFutbol" width="320">

# TuFutbol

TuFutbol es una aplicación web desarrollada en Go diseñada para llevar el registro individual de tus partidos de fútbol y gestionar tus estadísticas personales de rendimiento.

Desde el TP5 la aplicación se **renderiza del lado del servidor**: las páginas HTML las genera el servidor Go con componentes [templ](https://templ.guide/) a partir de los datos que obtiene de PostgreSQL con [sqlc](https://docs.sqlc.dev/). Ya no hay API JSON ni JavaScript en el navegador: los formularios se envían con `POST` y el servidor responde con una redirección (patrón Post/Redirect/Get).

---

## Requisitos previos
* **Go** 1.26 o superior
* **Docker** con **Docker Compose v2** (comando `docker compose`)
* **sqlc**
* **make**, **bash** y **curl**
* Un navegador web
* Conexión a internet la primera vez (para descargar las dependencias de Go y el generador de templ)

No hace falta instalar templ: `make` lo ejecuta con `go run` en la misma versión que figura en `go.mod`.

Los puertos **8080** (aplicación) y **5432** (PostgreSQL) tienen que estar libres. Si alguno está ocupado, se puede cambiar en el archivo `.env` (`API_PORT` y `DB_PORT`).

## Instrucciones de ejecución
1. Clonar el repositorio, pararse en la raíz y ubicarse en el branch `tp5`:

   ```bash
   git clone https://github.com/nicorimoldi/TuFutbol
   cd TuFutbol
   git checkout tp5
   ```

2. Iniciar Docker si no se encuentra activo:

   ```bash
   systemctl start docker
   ```

3. Ejecutar la aplicación:

   ```bash
   make run-app
   ```

   Cuando aparece `TuFutbol corriendo en http://localhost:8080`, abrir esa dirección en el navegador (si se cambió `API_PORT` en `.env`, usar ese puerto). Desde ahí se cargan partidos con el formulario, se ven en la lista (con filtros por resultado y el puntaje coloreado según el rendimiento) y se editan o borran con los botones **Editar** y **Borrar** de cada partido. Cada acción recarga la página completa.

   Para terminar, presionar **Ctrl+C**: se detiene el servidor y se borran los contenedores y los datos cargados.

4. (Opcional) Ejecutar las pruebas de la aplicación web y los tests de Go:

   ```bash
   make test-web
   make test
   ```

## Funcionamiento de `make run-app`

1. **Tareas previas:**
   * Creación de `.env` a partir de `.env.example`, si no existe.
   * Generación de código: `sqlc generate` (acceso a datos, en `db/sqlc/`) y `templ generate` (vistas, archivos `views/*_templ.go`).
   * Compilación del servidor (`go build -o bin/tufutbol .`).
   * Borrado de contenedores y volúmenes anteriores (`docker compose down -v`).
   * Arranque del contenedor de PostgreSQL y espera a que esté listo para recibir conexiones (`docker compose up -d --build --wait db`). La base crea las tablas al iniciarse.

2. **Ejecución de la aplicación:** se ejecuta el binario compilado, conectado a la base del contenedor, en http://localhost:8080. Queda corriendo hasta que se presiona Ctrl+C.

3. **Tareas posteriores:** al terminar (con Ctrl+C o si algo falla) se borran los contenedores, la red y el volumen de la base (`docker compose down -v`).

## Otros comandos

| Comando | Descripción |
|---|---|
| `make run-app` | Ejecuta la aplicación con todas las tareas previas y posteriores |
| `make test-web` | Levanta todo en Docker, prueba cada página y formulario con `curl` ([`requests.sh`](requests.sh)) y borra todo |
| `make test` | Tests de Go (capa de datos contra PostgreSQL en Docker, validación del formulario y componentes templ) |
| `make up` / `make down` | Levanta la base y la aplicación en Docker en segundo plano / las detiene y borra |
| `make generate` | Genera el código de sqlc y de templ |
| `make build` | Genera el código y compila el servidor |
| `make clean` | Borra el binario y el código generado |

## Documentación

La documentación del proyecto se encuentra en [`docs/DOCUMENTACION.md`](docs/DOCUMENTACION.md).
