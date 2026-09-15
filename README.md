# TuFutbol

TuFutbol es una aplicación web desarrollada en Go diseñada para llevar el registro individual de tus partidos de fútbol y gestionar tus estadísticas personales de rendimiento.

---

## Requisitos previos
* **Go** instalado en el sistema
* **Docker** instalado en el sistema
* **sqlc** instalado en el sistema

* El puerto **5432** del host tiene que estar libre. Si hay un PostgreSQL corriendo localmente, detenelo o cambiá `DB_PORT` en `test.sh`.

## Instrucciones de ejecución
1. Clonar el repositorio, pararse en la raíz y ubicarse en el branch `tp2`:

   ```bash
   git clone https://github.com/nicorimoldi/TuFutbol
   cd TuFutbol
   git checkout tp2
   ```

2. Iniciar Docker si no se encuentra activo:

   ```bash
   systemctl start docker
   ```
   
3. Ejecutar los tests mediante cualquiera de los siguientes comandos equivalentes:

   ```bash
   make test
   ```

   ```bash
   ./test.sh
   ```

## Funcionamiento del script `test.sh`

1. **Tareas previas:**
   * Limpieza de contenedores y volúmenes previos.
   * Verificación de compilación del proyecto en Go.
   * Creación de volumen dedicado y arranque del contenedor PostgreSQL.
   * Creación de tablas e inicialización de la base de datos.

2. **Ejecución de tests:**
   * Ejecución de los tests utilizando el paquete oficial `testing` de Go.

3. **Tareas posteriores:**
   * Eliminación del contenedor y volúmenes asociados, independientemente de si las pruebas pasan o fallan.

## Documentación

La documentación del proyecto se encuentra en [`docs/DOCUMENTACION.md`](docs/DOCUMENTACION.md).
