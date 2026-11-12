package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	db "TuFutbol/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib" // driver de PostgreSQL
	_ "time/tzdata"                    // zonas horarias incluidas en el binario (para la variable TZ)
)

// queries es el repositorio generado por sqlc. Lo usan todos los handlers.
var queries *db.Queries

func main() {
	// 1. Conexión con la base de datos. La URL llega en la variable de entorno
	//    DATABASE_URL: la arma el Makefile (make run-app) o docker-compose.yml
	//    (make up) con los datos de .env.
	conn, err := sql.Open("pgx", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	if err := conn.Ping(); err != nil {
		log.Fatal("No se pudo conectar a la base de datos: ", err)
	}

	// 2. Repositorio generado por sqlc
	queries = db.New(conn)

	// 3. Rutas: cada combinación de método + ruta ejecuta un handler (handlers.go).
	//    Todas devuelven HTML generado con templ (carpeta views) o una redirección.
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", paginaPrincipal)                     // lista + formulario
	mux.HandleFunc("POST /partidos", crearPartido)                  // envío del formulario
	mux.HandleFunc("GET /partidos/{id}/editar", editarPartido)      // formulario con los datos del partido
	mux.HandleFunc("POST /partidos/{id}/editar", actualizarPartido) // envío del formulario de edición
	mux.HandleFunc("POST /partidos/{id}/borrar", borrarPartido)     // botón Borrar
	mux.HandleFunc("/", noEncontrada)                               // cualquier otra ruta

	// Archivos estáticos (styles.css, favicon e imágenes) de la carpeta static/
	estaticos := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", estaticos))

	// 4. Servidor HTTP. El puerto es 8080, salvo que se indique otro en la variable PORT.
	puerto := os.Getenv("PORT")
	if puerto == "" {
		puerto = "8080"
	}
	log.Printf("Servidor escuchando en http://localhost:%s", puerto)
	log.Fatal(http.ListenAndServe(":"+puerto, mux))
}
