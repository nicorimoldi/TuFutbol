package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	db "TuFutbol/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib" // driver de PostgreSQL
)

// queries es el repositorio generado por sqlc. Lo usan todos los handlers.
var queries *db.Queries

func main() {
	// 1. Conexión con la base de datos. La URL llega en la variable de entorno
	//    DATABASE_URL, que docker-compose.yml arma con los datos de .env.
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

	// 3. Rutas: cada combinación de método + ruta ejecuta un handler (handlers.go)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /partidos", createPartido)
	mux.HandleFunc("GET /partidos", listPartidos)
	mux.HandleFunc("GET /partidos/{id}", getPartido)
	mux.HandleFunc("PUT /partidos/{id}", updatePartido)
	mux.HandleFunc("DELETE /partidos/{id}", deletePartido)
	mux.Handle("GET /", http.FileServer(http.Dir("static"))) // página de presentación

	// 4. Servidor HTTP
	log.Println("Servidor escuchando en http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
