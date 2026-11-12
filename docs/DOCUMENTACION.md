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
Las consultas se encuentran en [`db/queries/partidos.sql`](../db/queries/partidos.sql) y no cambiaron respecto del TP4:

* **Create (`CreatePartido`):** Inserción del registro del partido y estadísticas iniciales. Devuelve el partido creado (`:one` con `RETURNING *`).
* **Get (`GetPartido`):** Búsqueda por clave primaria (`:one`).
* **List (`ListPartidos`):** Obtención por fecha descendente de todos los partidos del usuario (`:many`).
* **Update (`UpdatePartido`):** Modificación de atributos y estadísticas de un encuentro existente. Devuelve el partido actualizado (`:one` con `RETURNING *`).
* **Delete (`DeletePartido`):** Eliminación de partido por clave primaria. Devuelve la cantidad de filas eliminadas (`:execrows`).

Que Create devuelva el partido permite redirigir a la página principal con su id para resaltarlo, y que Update y Delete informen si el partido existía permite responder `404` si el id no existe.

El motor de base de datos y la ruta del código generado se configuran en [`sqlc.yaml`](../sqlc.yaml).

---

## 3. Renderizado del lado del servidor (templ)

En el TP5 se reemplazó la arquitectura **API REST + cliente JavaScript** del TP4 por una aplicación **renderizada en el servidor**: el servidor Go arma el HTML completo de cada página con componentes [templ](https://templ.guide/) y el navegador solo lo muestra. Ya no hay handlers que devuelvan JSON ni archivos `.js`.

### Componentes (`views/`)

El `index.html` del TP4 se dividió en componentes templ. Cada archivo `.templ` se compila a código Go (`*_templ.go`) con `templ generate`; ese código generado no se sube al repo (está en `.gitignore`) y lo genera `make generate`.

| Archivo | Componente | Qué renderiza |
|---|---|---|
| [`layout.templ`](../views/layout.templ) | `Layout(titulo)` | Estructura base: `<!DOCTYPE html>`, `<head>` con el link a la hoja de estilos, la cabecera y el pie. Recibe el contenido de la página como hijo (`{ children... }`). |
| [`partido_list.templ`](../views/partido_list.templ) | `PartidoList(partidos []db.Partido, estado)` | Recibe el slice de partidos que devuelve `ListPartidos` (el struct generado por sqlc) y arma el resumen, los filtros, la leyenda y un `<li>` por partido. Sin partidos, muestra el dibujo del arco y la invitación a cargar el primero. |
| [`partido_form.templ`](../views/partido_form.templ) | `PartidoForm(form)` | El formulario para crear un partido (`POST /partidos`) o editarlo (`POST /partidos/{id}/editar`), con los mensajes de éxito o de error. |
| [`index.templ`](../views/index.templ) | `IndexPage(titulo, partidos, estado, form, aviso)` | Página principal: usa `Layout` y adentro renderiza `PartidoForm` y `PartidoList`. |
| [`layout.templ`](../views/layout.templ) | `PaginaError(titulo, mensaje)` | Página de error (`404`, `500`) con el mismo diseño y un enlace para volver. |
| [`datos.go`](../views/datos.go) | — | Los structs que reciben los componentes (`FormPartido`, `EstadoLista`) y las funciones que antes estaban en `app.js`: nivel del puntaje, color del resultado, insignias, resumen, últimos 5 y filtros. |

**Hoja de estilos:** la consigna sugiere un link a una hoja como Pico.css; se usa la hoja propia del TP4 ([`static/styles.css`](../static/styles.css)), que ya resolvía el diseño completo de la página, servida por el mismo servidor en `/static/`.

**Seguridad:** templ escapa automáticamente todo lo que se inserta con `{ ... }`, así que los textos que escribe el usuario (rival, resultado, posición) nunca se interpretan como HTML. En el TP4 esto se hacía a mano con `escaparHTML` en JavaScript.

### Servidor (`main.go` y `handlers.go`)

* **`main.go`:** Conexión con la base (variable `DATABASE_URL`), repositorio de sqlc, registro de rutas y servidor HTTP en el puerto 8080 (o el de la variable `PORT`).
* **`handlers.go`:** Un handler por ruta. Cada uno obtiene los datos con sqlc, llama al componente templ y lo renderiza en el `http.ResponseWriter` (`componente.Render(r.Context(), w)`).

| Método | Ruta | Descripción | Respuesta |
|---|---|---|---|
| `GET` | `/` | Lista de partidos (`ListPartidos`) y formulario vacío | `200` + HTML |
| `POST` | `/partidos` | Envío del formulario: crea el partido (`CreatePartido`) | `303` → `/?guardado={id}`, `422` + formulario con el error |
| `GET` | `/partidos/{id}/editar` | Página principal con el formulario completo con los datos del partido (`GetPartido`) | `200` + HTML, `404` si no existe |
| `POST` | `/partidos/{id}/editar` | Envío del formulario de edición (`UpdatePartido`) | `303` → `/?editado={id}`, `422` o `404` |
| `POST` | `/partidos/{id}/borrar` | Botón **Borrar** (`DeletePartido`) | `303` → `/?borrado={rival}`, `404` si no existe |
| `GET` | `/static/...` | Hoja de estilos, favicon e imágenes | `200` |

Los formularios HTML solo pueden enviar `GET` y `POST`, por eso editar y borrar usan `POST` a rutas propias en lugar de `PUT` y `DELETE`.

**Parámetros de `GET /`:** `resultado=ganado|empatado|perdido` filtra la lista; `guardado={id}` y `editado={id}` resaltan el partido y muestran el mensaje de éxito; `borrado={rival}` muestra el aviso de partido borrado. Así el estado de la página queda en la URL y se mantiene al recargarla.

### Patrón Post/Redirect/Get (PRG)

1. El navegador envía el formulario con `POST /partidos` (recarga completa, sin JavaScript).
2. El handler parsea los datos con `r.ParseForm()`, los valida y los guarda con `CreatePartido`.
3. Responde con `http.Redirect(w, r, "/?guardado={id}", http.StatusSeeOther)`.
4. El navegador pide `GET /?guardado={id}` y el handler principal muestra la lista actualizada, con el partido nuevo resaltado y el mensaje "Partido contra … guardado.".

Como la última petición es un `GET`, recargar la página no vuelve a enviar el formulario ni crea el partido dos veces. Editar y borrar siguen el mismo patrón.

Si los datos no son válidos no se redirige: se responde `422` con la página completa, el formulario con lo que se había escrito y el mensaje del error.

**Validaciones** (`validarPartido`, igual que en el TP4): fecha con formato `AAAA-MM-DD`, tipo de cancha mayor a 0 (sin restringir el tamaño de la cancha), rival y resultado obligatorios, goles y asistencias enteros no negativos, puntaje entre 1 y 10, y textos que no superen el largo de su columna en la base. Como el formulario envía todo como texto, la validación también revisa que los números sean números. Los inputs además validan en el navegador (campos obligatorios, mínimos y máximos, largo de los textos), pero la validación definitiva es la del servidor.

---

## 4. Estrategia de Pruebas y Aislamiento

* **Capa de datos** ([`Partido_test.go`](../Partido_test.go)): valida el ciclo de vida completo del CRUD contra PostgreSQL. Cada ejecución opera adentro de una transacción (`db.BeginTx`) que termina con `defer tx.Rollback()`, así los tests son idempotentes y no dejan datos en la base. La base corre en un contenedor que se crea y destruye para la prueba. Si no está definida `DATABASE_URL`, este test se omite.
* **Formulario** ([`handlers_test.go`](../handlers_test.go)): lectura y validación del formulario (`leerFormulario` / `validarPartido`) con casos válidos e inválidos. No usa la base.
* **Componentes templ** ([`views/views_test.go`](../views/views_test.go)): renderiza las vistas con datos de ejemplo y revisa el HTML (resumen, filtros, formulario de edición, lista vacía y que los textos del usuario se escapen). No usa la base.
* **Aplicación web completa** ([`requests.sh`](../requests.sh)): con la aplicación corriendo, pide cada página y envía los formularios con `curl` como lo haría el navegador. Verifica el código HTTP, la redirección `303` del patrón PRG y que el HTML tenga lo esperado (incluido que un rival con HTML aparezca escapado).

---

## 5. Automatización del Entorno (`Makefile` / Docker Compose)

Todos los comandos están en el [`Makefile`](../Makefile). La base de datos y la aplicación se definen como servicios en [`docker-compose.yml`](../docker-compose.yml).

### Configuración (`.env`)

La configuración está en el archivo `.env`, que no se sube al repo. En el repo está [`.env.example`](../.env.example) y `make` lo copia a `.env` automáticamente si no existe. Para cambiar un valor (por ejemplo, un puerto ocupado) se edita `.env`.

| Variable | Descripción | Valor por defecto |
|---|---|---|
| `POSTGRES_USER` | Usuario de la base de datos | `tufutbol` |
| `POSTGRES_PASSWORD` | Contraseña de la base de datos | `tufutbol` |
| `POSTGRES_DB` | Nombre de la base de datos | `tufutbol` |
| `DB_PORT` | Puerto de PostgreSQL en esta máquina | `5432` |
| `API_PORT` | Puerto de la aplicación web en esta máquina (el nombre viene del TP4) | `8080` |

Docker Compose lee `.env` para configurar la base y armar la `DATABASE_URL` del servicio `api`, y el `Makefile` lo lee para armar la URL de la base que usan `make run-app` y `make test`.

### Comandos

| Comando | Descripción |
|---|---|
| `make run-app` | Ejecuta la aplicación con todas las tareas previas y posteriores (ver abajo) |
| `make test-web` | Prueba completa de la aplicación web: levanta todo en Docker, ejecuta `requests.sh` y borra todo |
| `make test` | Tests de Go contra PostgreSQL en Docker |
| `make up` | Levanta la base y la aplicación en Docker, en segundo plano (http://localhost:8080) |
| `make down` | Detiene y borra los contenedores y el volumen de la base |
| `make logs` | Muestra los logs de la aplicación cuando corre en Docker |
| `make generate` | Genera el código de sqlc (`db/sqlc/`) y de templ (`views/*_templ.go`) |
| `make build` | Genera el código y compila el servidor en `bin/tufutbol` |
| `make clean` | Borra el binario y el código generado |

**templ sin instalarlo:** `make generate` ejecuta `go run github.com/a-h/templ/cmd/templ@<versión> generate`, tomando la versión de `go.mod` (`go list -m`). Así el generador y la librería de templ siempre coinciden y no hace falta instalar nada además de Go.

### Funcionamiento de `make run-app`

1. **Tareas previas:**
   * Creación de `.env` si no existe.
   * `sqlc generate` y `templ generate`.
   * Compilación del servidor (`go build -o bin/tufutbol .`).
   * Limpieza de contenedores y volúmenes anteriores (`docker compose down -v`).
   * Arranque de PostgreSQL con `docker compose up -d --build --wait db`. La imagen de la base incluye `schema.sql` ([`db/Dockerfile`](../db/Dockerfile)), que se aplica automáticamente al crearse la base, y `--wait` espera a que la base acepte conexiones (healthcheck con `pg_isready`).
2. **Ejecución:** el binario compilado corre en esta máquina, conectado a la base del contenedor (`DATABASE_URL` y `PORT` se arman con los valores de `.env`). Queda en primer plano hasta que se presiona Ctrl+C.
3. **Tareas posteriores:** un `trap` de bash ejecuta `docker compose down -v` al terminar, tanto con Ctrl+C como si algún paso falla (por ejemplo, un puerto ocupado), así nunca quedan contenedores ni volúmenes. Durante la limpieza se ignoran los Ctrl+C y `docker compose down` corre en otra sesión (`setsid`), así apretar Ctrl+C varias veces no la corta a la mitad.

`make test-web` y `make test` siguen la misma idea: limpian, levantan los contenedores con `--wait`, ejecutan las pruebas y borran todo aunque fallen (y en ese caso muestran el log de los contenedores). `make test-web` levanta también la aplicación dentro de Docker (servicio `api`, cuya imagen se construye con el [`Dockerfile`](../Dockerfile)), así se prueba la imagen.

---

## 6. Interfaz de Usuario

La interfaz es la misma del TP4, pero ahora el HTML lo genera el servidor (sección 3) y cada acción recarga la página completa. La carpeta [`static/`](../static) solo tiene la hoja de estilos y las imágenes.

* **Logo:** [`favicon.svg`](../static/favicon.svg) (ícono de la pestaña, con líneas más gruesas para que se vea en tamaño chico) e [`img/isotipo-negativo.svg`](../static/img/isotipo-negativo.svg) (versión blanca para la cabecera verde). El logo completo con el nombre está en [`docs/logo.svg`](logo.svg) (lo usa el README). Representa la mitad de una cancha con el punto central en amarillo: el jugador en el centro de su propio registro.

### Del TP4 (JavaScript) al TP5 (servidor)

| Funcionalidad | TP4 (`app.js` + API) | TP5 (templ + formularios) |
|---|---|---|
| Ver la lista | `fetch` a `GET /partidos` y armado del HTML en el navegador | `GET /` devuelve la página completa con la lista |
| Cargar un partido | `fetch` con `POST` y JSON, sin recargar | Formulario HTML con `POST /partidos` + redirección (PRG) |
| Editar | El formulario se completaba desde JavaScript; `PUT /partidos/{id}` | Enlace **Editar** a `/partidos/{id}/editar` (formulario completo desde el servidor); envío con `POST` |
| Borrar | `DELETE /partidos/{id}` | Botón **Borrar** dentro de un formulario `POST /partidos/{id}/borrar` |
| Aviso de borrado | Aviso de 8 segundos ocultado con `setTimeout` | El mismo aviso, ocultado a los 8 segundos con una animación de CSS |
| Filtros | Filtraban en el navegador la lista ya descargada | Enlaces a `/?resultado=ganado` (el servidor filtra) |
| Partido nuevo resaltado y mensajes | Estado en variables de JavaScript | Parámetros de la URL (`?guardado=`, `?editado=`, `?borrado=`) |
| Errores de validación | Mensaje del JSON de error de la API | La página vuelve con el formulario completo y el mensaje (`422`) |
| Fecha de hoy en el formulario | Calculada en el navegador | La pone el servidor (en Docker, con la zona horaria de Argentina, variable `TZ`) |
| Escapado de textos | Función `escaparHTML` | Automático en templ |

**Lo que se quitó** porque depende de JavaScript en el navegador: el botón **Deshacer** del aviso de borrado y el color del campo Puntaje mientras se escribe (ahora el campo toma el color de su nivel cuando ya tiene un valor, por ejemplo al editar).

### Detalles de la lista

El resumen muestra los totales, el puntaje promedio y los últimos 5 resultados como G / E / P (el de la derecha es el último partido), como en las tablas de posiciones. Los partidos con 3, 4 o 5+ goles llevan la insignia Hat-trick, Póker o Manito. Cada filtro muestra cuántos partidos hay y el elegido se marca con `aria-current`. Todo esto se calcula en Go ([`views/datos.go`](../views/datos.go)).

### Estilos (`styles.css`)

* **CSS Grid** para el layout principal: dos columnas (formulario a la izquierda y lista a la derecha) que pasan a una sola columna en pantallas de menos de 860px. Cada partido de la lista también es una grilla (fecha | datos | estadísticas | botones).
* **Flexbox** para las filas de campos del formulario, el resumen, los filtros y la leyenda.
* En pantallas grandes el formulario queda fijo (`position: sticky`) al bajar por la lista.
* La página entera es media cancha vista desde arriba:
  * **Cabecera:** el pasto a franjas (`repeating-linear-gradient`) con la línea de mitad de cancha y el círculo y el punto centrales. En pantallas angostas se oculta el círculo para que no tape el título.
  * **Fondo:** sigue las franjas en verde claro.
  * **Líneas laterales:** recorren toda la página por los dos costados, desde la cabecera hasta el pie. Se dibujan como capas de `background` (un `linear-gradient` de un solo color, de 3px de ancho) en la cabecera, en el `body` y en el pie, todas a la misma distancia del borde (`--lateral`), así forman una sola línea continua.
  * **Pie:** cierra con la línea de fondo, el arco con la red detrás de ella (`::after`) y el área grande, el área chica, la medialuna y el punto penal (SVG).
* El resultado se colorea según sea ganado, empatado o perdido.
* **Puntaje según el rendimiento:** El puntaje de cada partido y el promedio del resumen se colorean en cuatro niveles: flojo (1 a 4), regular (5 y 6), bueno (7 y 8) y figura (9 y 10), con una leyenda arriba de la lista. Los colores van de claro a oscuro a medida que mejora el puntaje, así se distinguen aunque no se perciban bien los colores, y los lectores de pantalla leen el nivel junto al número.
* **Cambios del TP5:** los filtros, **Editar** y **Cancelar** ahora son enlaces y **Borrar** es el botón de un formulario, así que se agregaron los estilos para que se vean igual que antes; el aviso de borrado se oculta solo con una animación de CSS (`animation-delay` de 8 segundos), y se agregó el estilo de la página de error.
* Diseño adaptable a celulares, foco visible para navegar con teclado y sin animaciones si el sistema tiene activada la opción de reducir movimiento.
