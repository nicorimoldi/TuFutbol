# TuFutbol

TuFutbol es una aplicación web desarrollada en Go diseñada para llevar el registro individual de tus partidos de fútbol y gestionar tus estadísticas personales de rendimiento.

---

## Requisitos previos
* **Go** 1.26 o superior
* **Docker** con **Docker Compose v2** (comando `docker compose`)
* **sqlc**
* **make** y **curl**

* Los puertos **8080** (API) y **5432** (PostgreSQL) tienen que estar libres. Si alguno está ocupado, se puede cambiar en el archivo `.env` (`API_PORT` y `DB_PORT`).

## Instrucciones de ejecución
1. Clonar el repositorio, pararse en la raíz y ubicarse en el branch `tp3`:

   ```bash
   git clone https://github.com/nicorimoldi/TuFutbol
   cd TuFutbol
   git checkout tp3
   ```

2. Iniciar Docker si no se encuentra activo:

   ```bash
   systemctl start docker
   ```

3. Ejecutar las pruebas de la API:

   ```bash
   make test-api
   ```

4. (Opcional) Ejecutar los tests de Go de la capa de datos:

   ```bash
   make test
   ```

## Funcionamiento de `make test-api`

1. **Tareas previas:**
   * Creación de `.env` a partir de `.env.example`, si no existe.
   * Ejecución de `sqlc generate` y compilación del proyecto (`go build`).
   * Limpieza de contenedores y volúmenes anteriores (`docker compose down -v`).
   * Construcción de la imagen de la API y arranque de la base y la API (`docker compose up -d --build --wait`). La base crea las tablas al iniciarse, y `--wait` espera a que la base y la API estén listas.

2. **Prueba de cada uno de los endpoints:** [`requests.sh`](requests.sh) hace 10 peticiones con `curl` a los 5 endpoints (casos válidos y de error), muestra cada respuesta y verifica el código HTTP.

3. **Tareas posteriores:** eliminación de los contenedores, la red y el volumen (`docker compose down -v`), aunque las pruebas fallen.

Si todo funciona, termina con `Todas las pruebas pasaron`. Si algo falla, muestra el log de los contenedores para ver el motivo y `make` termina con error.


## Documentación

La documentación del proyecto se encuentra en [`docs/DOCUMENTACION.md`](docs/DOCUMENTACION.md).
