package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	db "TuFutbol/db/sqlc"
)

// partidoJSON es el formato en que la API recibe y devuelve un partido.
// Se usa en lugar de db.Partido (el struct de sqlc) para que la fecha sea
// "AAAA-MM-DD" y la posición sea un texto o null.
type partidoJSON struct {
	IDPartido          int32   `json:"id_partido"`
	Fecha              string  `json:"fecha"`
	TipoCancha         int32   `json:"tipo_cancha"`
	Rival              string  `json:"rival"`
	Resultado          string  `json:"resultado"`
	Posicion           *string `json:"posicion"`
	GolesPropios       int32   `json:"goles_propios"`
	AsistenciasPropias int32   `json:"asistencias_propias"`
	PuntajePropio      int32   `json:"puntaje_propio"`
}

// ---------------------------------------------------------------------------
// Handlers CRUD
// ---------------------------------------------------------------------------

// POST /partidos
func createPartido(w http.ResponseWriter, r *http.Request) {
	datos, err := leerPartido(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}

	partido, err := queries.CreatePartido(r.Context(), datos)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusCreated, aJSON(partido))
}

// GET /partidos
func listPartidos(w http.ResponseWriter, r *http.Request) {
	partidos, err := queries.ListPartidos(r.Context())
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	lista := []partidoJSON{} // si no hay partidos, se devuelve [] y no null
	for _, p := range partidos {
		lista = append(lista, aJSON(p))
	}
	responderJSON(w, http.StatusOK, lista)
}

// GET /partidos/{id}
func getPartido(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, "el id debe ser un número")
		return
	}

	partido, err := queries.GetPartido(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		responderError(w, http.StatusNotFound, "no existe el partido")
		return
	}
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, aJSON(partido))
}

// PUT /partidos/{id}
func updatePartido(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, "el id debe ser un número")
		return
	}
	datos, err := leerPartido(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, err.Error())
		return
	}

	partido, err := queries.UpdatePartido(r.Context(), db.UpdatePartidoParams{
		IDPartido:          id,
		Fecha:              datos.Fecha,
		TipoCancha:         datos.TipoCancha,
		Rival:              datos.Rival,
		Resultado:          datos.Resultado,
		Posicion:           datos.Posicion,
		GolesPropios:       datos.GolesPropios,
		AsistenciasPropias: datos.AsistenciasPropias,
		PuntajePropio:      datos.PuntajePropio,
	})
	if errors.Is(err, sql.ErrNoRows) { // UPDATE ... RETURNING no devolvió filas: el id no existe
		responderError(w, http.StatusNotFound, "no existe el partido")
		return
	}
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}

	responderJSON(w, http.StatusOK, aJSON(partido))
}

// DELETE /partidos/{id}
func deletePartido(w http.ResponseWriter, r *http.Request) {
	id, err := leerID(r)
	if err != nil {
		responderError(w, http.StatusBadRequest, "el id debe ser un número")
		return
	}

	filasBorradas, err := queries.DeletePartido(r.Context(), id)
	if err != nil {
		responderError(w, http.StatusInternalServerError, err.Error())
		return
	}
	if filasBorradas == 0 {
		responderError(w, http.StatusNotFound, "no existe el partido")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Validación (lógica de negocio)
// ---------------------------------------------------------------------------

// validarPartido revisa las reglas de negocio de un partido. Los largos máximos
// son los de las columnas varchar de db/schema/schema.sql.
func validarPartido(p partidoJSON) error {
	if _, err := time.Parse("2006-01-02", p.Fecha); err != nil {
		return errors.New("la fecha es obligatoria y debe tener el formato AAAA-MM-DD")
	}
	if p.TipoCancha <= 0 {
		return errors.New("tipo_cancha debe ser mayor a 0")
	}
	if strings.TrimSpace(p.Rival) == "" || utf8.RuneCountInString(p.Rival) > 50 {
		return errors.New("el rival es obligatorio y puede tener hasta 50 caracteres")
	}
	if strings.TrimSpace(p.Resultado) == "" || utf8.RuneCountInString(p.Resultado) > 20 {
		return errors.New("el resultado es obligatorio y puede tener hasta 20 caracteres")
	}
	if p.Posicion != nil && utf8.RuneCountInString(*p.Posicion) > 10 {
		return errors.New("la posición puede tener hasta 10 caracteres")
	}
	if p.GolesPropios < 0 || p.AsistenciasPropias < 0 {
		return errors.New("los goles y las asistencias no pueden ser negativos")
	}
	if p.PuntajePropio < 1 || p.PuntajePropio > 10 {
		return errors.New("el puntaje debe estar entre 1 y 10")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Funciones auxiliares
// ---------------------------------------------------------------------------

// leerPartido lee el JSON del body, lo valida y lo convierte a los tipos que usa sqlc.
func leerPartido(r *http.Request) (db.CreatePartidoParams, error) {
	var p partidoJSON
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		return db.CreatePartidoParams{}, errors.New("el body debe ser un JSON válido")
	}
	if err := validarPartido(p); err != nil {
		return db.CreatePartidoParams{}, err
	}

	fecha, _ := time.Parse("2006-01-02", p.Fecha) // ya se validó el formato
	posicion := sql.NullString{}                  // NULL si no se envió
	if p.Posicion != nil {
		posicion = sql.NullString{String: *p.Posicion, Valid: true}
	}

	return db.CreatePartidoParams{
		Fecha:              fecha,
		TipoCancha:         p.TipoCancha,
		Rival:              p.Rival,
		Resultado:          p.Resultado,
		Posicion:           posicion,
		GolesPropios:       p.GolesPropios,
		AsistenciasPropias: p.AsistenciasPropias,
		PuntajePropio:      p.PuntajePropio,
	}, nil
}

// aJSON convierte un partido de la base (struct de sqlc) al formato de la API.
func aJSON(p db.Partido) partidoJSON {
	var posicion *string // null si en la base es NULL
	if p.Posicion.Valid {
		posicion = &p.Posicion.String
	}
	return partidoJSON{
		IDPartido:          p.IDPartido,
		Fecha:              p.Fecha.Format("2006-01-02"),
		TipoCancha:         p.TipoCancha,
		Rival:              p.Rival,
		Resultado:          p.Resultado,
		Posicion:           posicion,
		GolesPropios:       p.GolesPropios,
		AsistenciasPropias: p.AsistenciasPropias,
		PuntajePropio:      p.PuntajePropio,
	}
}

// leerID obtiene el {id} de la URL.
func leerID(r *http.Request) (int32, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	return int32(id), err
}

func responderJSON(w http.ResponseWriter, codigo int, datos any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(codigo)
	json.NewEncoder(w).Encode(datos)
}

func responderError(w http.ResponseWriter, codigo int, mensaje string) {
	responderJSON(w, codigo, map[string]string{"error": mensaje})
}
