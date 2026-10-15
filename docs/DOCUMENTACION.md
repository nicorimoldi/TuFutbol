# Documentación

---

## 1. Decisiones de Diseño y Modelo Relacional

### Entidad Principal: `partido`
La aplicación modela y registra el rendimiento individual del usuario en partidos de fútbol. 

El esquema formal se encuentra definido en [`db/schema/schema.sql`](../db/schema/schema.sql) y se estructura bajo las siguientes pautas:

* **Identificación:** Clave primaria que se autoincrementa (`SERIAL`)
* **Contexto del encuentro:**
  * `fecha`: Almacena el día del encuentro (`date`).
  * `tipo_cancha`, `rival`, `resultado`: Definen el marco del partido (modalidad fútbol 5/7/11 u otra, nombre del rival y marcador general).
* **Estadísticas individuales de rendimiento:**
  * `goles_propios`, `asistencias_propias`, `puntaje_propio`: Estadísticas estrictamente cuantitativas del usuario.
  * `posicion`: Campo opcional (`NULL` permitido) para contemplar los partidos donde el usuario no tomó una posición fija.

---

## 2. Capa de Acceso a Datos (sqlc)

Para interactuar con la base de datos se utiliza **sqlc** sobre PostgreSQL.

* Valida la sintaxis SQL en tiempo de compilación y genera código Go fuertemente tipado.
* El código Go resultante se regenera automáticamente sin requerir mantenimiento manual. No se sube al repo (`db/sqlc/` está en `.gitignore`): lo genera `make generate` antes de compilar o testear.

### Operaciones CRUD implementadas
Las consultas se encuentran en [`db/queries/partidos.sql`](../db/queries/partidos.sql):

* **Create (`CreatePartido`):** Inserción del registro del partido y estadísticas iniciales. Devuelve el partido creado (`:one` con `RETURNING *`).
* **Get (`GetPartido`):** Búsqueda por clave primaria (`:one`).
* **List (`ListPartidos`):** Obtención por fecha descendente de todos los partidos del usuario (`:many`).
* **Update (`UpdatePartido`):** Modificación de atributos y estadísticas de un encuentro existente. Devuelve el partido actualizado (`:one` con `RETURNING *`).
* **Delete (`DeletePartido`):** Eliminación de partido por clave primaria. Devuelve la cantidad de filas eliminadas (`:execrows`).

Que Create y Update devuelvan el partido, y que Delete devuelva las filas eliminadas, permite que la API responda con el partido o con `404` si el id no existe.

El motor de base de datos y la ruta del código generado se configuran en [`sqlc.yaml`](../sqlc.yaml).

---

## 3. API REST

La capa de datos se expone como una API REST:

* **`main.go`:** Establece la conexión con la base de datos (variable de entorno `DATABASE_URL`), crea el repositorio de sqlc, registra las rutas y levanta el servidor en el puerto 8080.
* **`handlers.go`:** Un handler por cada operación CRUD, más la validación de los datos (`validarPartido`).

| Método | Ruta | Descripción | Respuesta |
|---|---|---|---|
| `POST` | `/partidos` | Crea un partido | `201` + partido creado, `400` si los datos no son válidos |
| `GET` | `/partidos` | Lista todos los partidos | `200` + array de partidos |
| `GET` | `/partidos/{id}` | Obtiene un partido | `200` + partido, `404` si no existe |
| `PUT` | `/partidos/{id}` | Actualiza un partido | `200` + partido actualizado, `400` o `404` |
| `DELETE` | `/partidos/{id}` | Elimina un partido | `204`, `404` si no existe |

**Validaciones** (en `POST` y `PUT`): fecha con formato `AAAA-MM-DD`, `tipo_cancha` mayor a 0 (sin restringir el tamaño de la cancha), `rival` y `resultado` obligatorios, goles y asistencias no negativos, puntaje entre 1 y 10, y textos que no superen el largo de su columna en la base.

---

## 4. Estrategia de Pruebas y Aislamiento

Las pruebas automáticas implementadas en [`Partido_test.go`](../Partido_test.go) validan el ciclo de vida completo del CRUD bajo dos criterios de calidad:

1. **Aislamiento transaccional:** Cada ejecución de pruebas opera adentro de una transacción de base de datos (`db.BeginTx`) que ejecuta un `defer tx.Rollback()`. Esto hace que los tests sean idempotentes y no dejen datos basura en la base de datos.
2. **Entorno reproducible con Docker:** La base de datos corre dentro de un contenedor que se crea y destruye para la prueba, así no hace falta configurar un PostgreSQL propio en la computadora.

Además, la API se prueba con [`requests.sh`](../requests.sh), que hace 10 peticiones con `curl` a los 5 endpoints (casos válidos y de error) y verifica el código HTTP de cada respuesta.

---

## 5. Automatización del Entorno (`Makefile` / Docker Compose)

Todos los comandos están en el [`Makefile`](../Makefile). La base de datos y la API se definen como servicios en [`docker-compose.yml`](../docker-compose.yml).

### Configuración (`.env`)

La configuración está en el archivo `.env`, que no se sube al repo. En el repo está [`.env.example`](../.env.example) y `make` lo copia a `.env` automáticamente si no existe. Para cambiar un valor (por ejemplo, un puerto ocupado) se edita `.env`.

| Variable | Descripción | Valor por defecto |
|---|---|---|
| `POSTGRES_USER` | Usuario de la base de datos | `tufutbol` |
| `POSTGRES_PASSWORD` | Contraseña de la base de datos | `tufutbol` |
| `POSTGRES_DB` | Nombre de la base de datos | `tufutbol` |
| `DB_PORT` | Puerto de PostgreSQL en esta máquina | `5432` |
| `API_PORT` | Puerto de la API en esta máquina | `8080` |

Docker Compose lee `.env` para configurar la base y armar la `DATABASE_URL` de la API, y el `Makefile` lo lee para armar la URL de la base que usa `make test`.

### Comandos

| Comando | Descripción |
|---|---|
| `make test-api` | Prueba completa de la API: levanta todo, prueba cada endpoint y borra todo |
| `make test` | Tests de Go de la capa de datos contra PostgreSQL en Docker |
| `make up` | Levanta la base y la API (http://localhost:8080) |
| `make down` | Detiene y borra los contenedores y el volumen de la base |
| `make logs` | Muestra los logs de la API |
| `make generate` | Genera el código de sqlc |
| `make build` | Compila el proyecto |

Para probar la API a mano: `make up`, hacer las peticiones con `curl` (o ejecutar `./requests.sh`) y al terminar `make down`.

### Funcionamiento de `make test-api`

El comando `make test-api` automatiza todo el proceso sin necesidad de pasos manuales:

1. **Preparación:** Creación de `.env` si no existe, ejecución de `sqlc generate`, compilación de Go y limpieza preventiva de contenedores y volúmenes previos.
2. **Levantamiento y espera de conexión:** `docker compose up --wait` levanta la base y la API, y espera a que ambas estén listas (la base mediante `pg_isready`).
3. **Carga del esquema:** La imagen de la base incluye `schema.sql` ([`db/Dockerfile`](../db/Dockerfile)), que se aplica automáticamente al crearse la base.
4. **Ejecución:** Prueba de cada endpoint con `requests.sh`.
5. **Limpieza final:** Eliminación de contenedores y volúmenes con `docker compose down -v`, ejecutándose incluso si las pruebas fallan.

El comando `make test` sigue los mismos pasos, pero levanta solo la base de datos y ejecuta los tests de Go mediante `go test -v ./...`.
