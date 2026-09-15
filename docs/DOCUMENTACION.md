# Documentación

---

## 1. Decisiones de Diseño y Modelo Relacional

### Entidad Principal: `partido`
La aplicación modela y registra el rendimiento individual del usuario en partidos de fútbol. 

El esquema formal se encuentra definido en [`db/schema/schema.sql`](../db/schema/schema.sql) y se estructura bajo las siguientes pautas:

* **Identificación:** Clave primaria que se autoincrementa (`SERIAL`)
* **Contexto del encuentro:**
  * `fecha`: Almacena el día del encuentro (`date`).
  * `tipo_cancha`, `rival`, `resultado`: Definen el marco del partido (modalidad fútbol 5/7/11, nombre del rival y marcador general).
* **Estadísticas individuales de rendimiento:**
  * `goles_propios`, `asistencias_propias`, `puntaje_propio`: Estadísticas estrictamente cuantitativas del usuario.
  * `posicion`: Campo opcional (`NULL` permitido) para contemplar los partidos donde el usuario no tomó una posición fija.

---

## 2. Capa de Acceso a Datos (sqlc)

Para interactuar con la base de datos se utiliza **sqlc** sobre PostgreSQL.

* Valida la sintaxis SQL en tiempo de compilación y genera código Go fuertemente tipado.
* El código Go resultante se regenera automáticamente sin requerir mantenimiento manual.

### Operaciones CRUD implementadas
Las consultas se encuentran en [`db/queries/users.sql`](../db/queries/users.sql):

* **Create (`CreatePartido`):** Inserción del registro del partido y estadísticas iniciales.
* **Get (`GetPartido`):** Búsqueda por clave primaria (`:one`).
* **List (`ListPartidos`):** Obtención por fecha descendente de todos los partidos del usuario (`:many`).
* **Update (`UpdatePartido`):** Modificación de atributos y estadísticas de un encuentro existente.
* **Delete (`DeletePartido`):** Eliminación de partido por clave primaria.

El motor de base de datos y la ruta del código generado se configuran en [`sqlc.yaml`](../sqlc.yaml).

---

## 3. Estrategia de Pruebas y Aislamiento

Las pruebas automáticas implementadas en [`Partido_test.go`](../Partido_test.go) validan el ciclo de vida completo del CRUD bajo dos criterios de calidad:

1. **Aislamiento transaccional:** Cada ejecución de pruebas opera adentro de una transacción de base de datos (`db.BeginTx`) que ejecuta un `defer tx.Rollback()`. Esto hace que los tests sean idempotentes y no dejen datos basura en la base de datos.
2. **Entorno reproducible con Docker:** La base de datos corre dentro de un contenedor que se crea y destruye para la prueba, así no hace falta configurar un PostgreSQL propio en la computadora.

---

## 4. Automatización del Entorno (`test.sh` / `Makefile`)

El script ejecutado mediante `make test` o `./test.sh` automatiza todo el proceso sin necesidad de pasos manuales:

1. **Preparación:** Limpieza preventiva de contenedores previos, ejecución de `sqlc generate`, compilación de Go y levantamiento del contenedor PostgreSQL.
2. **Espera de conexión:** Espera activa de conectividad mediante `pg_isready`.
3. **Carga del esquema:** Aplicación del archivo `schema.sql` sobre la base de pruebas.
4. **Ejecución:** Ejecución de los tests mediante `go test -v ./...`.
5. **Limpieza final:** Limpieza garantizada de contenedores y volúmenes mediante un `trap cleanup EXIT` de bash, ejecutándose incluso si las pruebas fallan.