package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	db "TuFutbol/db/sqlc"
	"TuFutbol/views"

	"github.com/a-h/templ"
)

// Los handlers responden con HTML generado por los componentes templ de la
// carpeta views. Los formularios siguen el patrón Post/Redirect/Get (PRG):
// si los datos son válidos, se guardan y se redirige a GET / con
// http.Redirect, así el navegador vuelve a pedir la página y se ve la lista
// actualizada (y recargar la página no reenvía el formulario).

const titulo = "TuFutbol"

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// GET /
// Muestra la lista de partidos y el formulario para cargar uno nuevo.
// Parámetros opcionales de la URL:
//   - resultado=ganado|empatado|perdido: filtra la lista
//   - guardado={id} o editado={id}: resalta el partido y muestra el mensaje de éxito
//   - borrado={rival}: muestra el aviso de partido borrado
func paginaPrincipal(w http.ResponseWriter, r *http.Request) {
	partidos, err := queries.ListPartidos(r.Context())
	if err != nil {
		errorInterno(w, r, err)
		return
	}

	parametros := r.URL.Query()
	estado := views.EstadoLista{Filtro: views.FiltroValido(parametros.Get("resultado"))}
	form := formularioVacio()

	if partido, ok := buscarPartido(partidos, parametros.Get("guardado")); ok {
		estado.Resaltado = partido.IDPartido
		form.Mensaje = fmt.Sprintf("Partido contra %s guardado.", partido.Rival)
	}
	if partido, ok := buscarPartido(partidos, parametros.Get("editado")); ok {
		estado.Resaltado = partido.IDPartido
		form.Mensaje = fmt.Sprintf("Cambios guardados en el partido contra %s.", partido.Rival)
	}

	aviso := ""
	if rival := parametros.Get("borrado"); rival != "" {
		aviso = fmt.Sprintf("Borraste el partido contra %s.", rival)
	}

	mostrarPagina(w, r, http.StatusOK, views.IndexPage(titulo, partidos, estado, form, aviso))
}

// POST /partidos
// Recibe el formulario, crea el partido y redirige a la página principal.
func crearPartido(w http.ResponseWriter, r *http.Request) {
	form, datos, err := leerFormulario(r)
	if err != nil {
		mostrarError(w, r, form, err)
		return
	}

	partido, err := queries.CreatePartido(r.Context(), datos)
	if err != nil {
		errorInterno(w, r, err)
		return
	}

	// 303 See Other: el navegador pide GET / (no repite el POST)
	http.Redirect(w, r, fmt.Sprintf("/?guardado=%d", partido.IDPartido), http.StatusSeeOther)
}

// GET /partidos/{id}/editar
// Muestra la página principal con el formulario completo con los datos del partido.
func editarPartido(w http.ResponseWriter, r *http.Request) {
	partido, ok := obtenerPartido(w, r)
	if !ok {
		return
	}
	partidos, err := queries.ListPartidos(r.Context())
	if err != nil {
		errorInterno(w, r, err)
		return
	}

	form := formularioDe(partido)
	estado := views.EstadoLista{Editando: partido.IDPartido}
	mostrarPagina(w, r, http.StatusOK, views.IndexPage("Editar partido · "+titulo, partidos, estado, form, ""))
}

// POST /partidos/{id}/editar
// Recibe el formulario de edición, actualiza el partido y redirige a la página principal.
func actualizarPartido(w http.ResponseWriter, r *http.Request) {
	partido, ok := obtenerPartido(w, r)
	if !ok {
		return
	}

	form, datos, err := leerFormulario(r)
	if err != nil {
		form.ID = partido.IDPartido
		form.RivalActual = partido.Rival
		mostrarError(w, r, form, err)
		return
	}

	_, err = queries.UpdatePartido(r.Context(), db.UpdatePartidoParams{
		IDPartido:          partido.IDPartido,
		Fecha:              datos.Fecha,
		TipoCancha:         datos.TipoCancha,
		Rival:              datos.Rival,
		Resultado:          datos.Resultado,
		Posicion:           datos.Posicion,
		GolesPropios:       datos.GolesPropios,
		AsistenciasPropias: datos.AsistenciasPropias,
		PuntajePropio:      datos.PuntajePropio,
	})
	if errors.Is(err, sql.ErrNoRows) { // lo borraron mientras se editaba
		noExistePartido(w, r)
		return
	}
	if err != nil {
		errorInterno(w, r, err)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/?editado=%d", partido.IDPartido), http.StatusSeeOther)
}

// POST /partidos/{id}/borrar
// Borra el partido y redirige a la página principal, que muestra el aviso.
func borrarPartido(w http.ResponseWriter, r *http.Request) {
	partido, ok := obtenerPartido(w, r)
	if !ok {
		return
	}

	filasBorradas, err := queries.DeletePartido(r.Context(), partido.IDPartido)
	if err != nil {
		errorInterno(w, r, err)
		return
	}
	if filasBorradas == 0 {
		noExistePartido(w, r)
		return
	}

	http.Redirect(w, r, "/?borrado="+url.QueryEscape(partido.Rival), http.StatusSeeOther)
}

// Cualquier ruta que no existe
func noEncontrada(w http.ResponseWriter, r *http.Request) {
	mostrarPagina(w, r, http.StatusNotFound, views.PaginaError("Página no encontrada", "La dirección que pediste no existe."))
}

// ---------------------------------------------------------------------------
// Formulario: lectura y validación (lógica de negocio)
// ---------------------------------------------------------------------------

// leerFormulario parsea los datos del formulario (r.ParseForm), los valida y
// los convierte a los tipos que usa sqlc. También devuelve los valores tal como
// se escribieron, para volver a mostrarlos en el formulario si hay un error.
func leerFormulario(r *http.Request) (views.FormPartido, db.CreatePartidoParams, error) {
	if err := r.ParseForm(); err != nil {
		return formularioVacio(), db.CreatePartidoParams{}, errors.New("No se pudieron leer los datos del formulario.")
	}

	form := views.FormPartido{
		Fecha:       strings.TrimSpace(r.PostForm.Get("fecha")),
		TipoCancha:  strings.TrimSpace(r.PostForm.Get("tipo_cancha")),
		Rival:       strings.TrimSpace(r.PostForm.Get("rival")),
		Resultado:   strings.TrimSpace(r.PostForm.Get("resultado")),
		Posicion:    strings.TrimSpace(r.PostForm.Get("posicion")),
		Goles:       strings.TrimSpace(r.PostForm.Get("goles_propios")),
		Asistencias: strings.TrimSpace(r.PostForm.Get("asistencias_propias")),
		Puntaje:     strings.TrimSpace(r.PostForm.Get("puntaje_propio")),
	}
	datos, err := validarPartido(form)
	return form, datos, err
}

// validarPartido revisa las reglas de negocio de un partido y convierte los
// textos del formulario a los tipos de la base. Los largos máximos son los de
// las columnas varchar de db/schema/schema.sql.
func validarPartido(f views.FormPartido) (db.CreatePartidoParams, error) {
	fecha, err := time.Parse("2006-01-02", f.Fecha)
	if err != nil {
		return db.CreatePartidoParams{}, errors.New("La fecha es obligatoria y debe tener el formato AAAA-MM-DD.")
	}
	tipoCancha, err := aNumero(f.TipoCancha)
	if err != nil || tipoCancha <= 0 {
		return db.CreatePartidoParams{}, errors.New("El tipo de cancha debe ser un número mayor a 0.")
	}
	if f.Rival == "" || utf8.RuneCountInString(f.Rival) > 50 {
		return db.CreatePartidoParams{}, errors.New("El rival es obligatorio y puede tener hasta 50 caracteres.")
	}
	if f.Resultado == "" || utf8.RuneCountInString(f.Resultado) > 20 {
		return db.CreatePartidoParams{}, errors.New("El resultado es obligatorio y puede tener hasta 20 caracteres.")
	}
	if utf8.RuneCountInString(f.Posicion) > 10 {
		return db.CreatePartidoParams{}, errors.New("La posición puede tener hasta 10 caracteres.")
	}
	goles, errGoles := aNumero(f.Goles)
	asistencias, errAsistencias := aNumero(f.Asistencias)
	if errGoles != nil || errAsistencias != nil || goles < 0 || asistencias < 0 {
		return db.CreatePartidoParams{}, errors.New("Los goles y las asistencias deben ser números enteros, sin negativos.")
	}
	puntaje, err := aNumero(f.Puntaje)
	if err != nil || puntaje < 1 || puntaje > 10 {
		return db.CreatePartidoParams{}, errors.New("El puntaje debe ser un número entero entre 1 y 10.")
	}

	return db.CreatePartidoParams{
		Fecha:              fecha,
		TipoCancha:         tipoCancha,
		Rival:              f.Rival,
		Resultado:          f.Resultado,
		Posicion:           sql.NullString{String: f.Posicion, Valid: f.Posicion != ""}, // vacía: NULL
		GolesPropios:       goles,
		AsistenciasPropias: asistencias,
		PuntajePropio:      puntaje,
	}, nil
}

// ---------------------------------------------------------------------------
// Funciones auxiliares
// ---------------------------------------------------------------------------

// mostrarPagina renderiza el componente templ en el http.ResponseWriter.
func mostrarPagina(w http.ResponseWriter, r *http.Request, codigo int, pagina templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(codigo)
	if err := pagina.Render(r.Context(), w); err != nil {
		log.Println("Error al renderizar la página:", err)
	}
}

// mostrarError vuelve a mostrar la página con el formulario completo con lo que
// se había escrito y el mensaje del error (422: los datos no son válidos).
func mostrarError(w http.ResponseWriter, r *http.Request, form views.FormPartido, err error) {
	partidos, errLista := queries.ListPartidos(r.Context())
	if errLista != nil {
		errorInterno(w, r, errLista)
		return
	}
	form.Error = err.Error()
	estado := views.EstadoLista{Editando: form.ID}
	tituloPagina := titulo
	if form.ID != 0 {
		tituloPagina = "Editar partido · " + titulo
	}
	mostrarPagina(w, r, http.StatusUnprocessableEntity, views.IndexPage(tituloPagina, partidos, estado, form, ""))
}

// obtenerPartido busca el partido del {id} de la URL. Si el id no es un número
// o el partido no existe, responde 404 y devuelve ok = false.
func obtenerPartido(w http.ResponseWriter, r *http.Request) (partido db.Partido, ok bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 32)
	if err != nil {
		noExistePartido(w, r)
		return db.Partido{}, false
	}
	partido, err = queries.GetPartido(r.Context(), int32(id))
	if errors.Is(err, sql.ErrNoRows) {
		noExistePartido(w, r)
		return db.Partido{}, false
	}
	if err != nil {
		errorInterno(w, r, err)
		return db.Partido{}, false
	}
	return partido, true
}

// buscarPartido busca en la lista el partido con el id indicado (texto de la URL).
func buscarPartido(partidos []db.Partido, idTexto string) (db.Partido, bool) {
	id, err := strconv.ParseInt(idTexto, 10, 32)
	if err != nil {
		return db.Partido{}, false
	}
	for _, p := range partidos {
		if p.IDPartido == int32(id) {
			return p, true
		}
	}
	return db.Partido{}, false
}

// formularioVacio es el formulario para cargar un partido nuevo: la fecha de
// hoy y 0 goles y asistencias.
func formularioVacio() views.FormPartido {
	return views.FormPartido{
		Fecha:       time.Now().Format("2006-01-02"),
		Goles:       "0",
		Asistencias: "0",
	}
}

// formularioDe completa el formulario con los datos de un partido de la base (para editarlo).
func formularioDe(p db.Partido) views.FormPartido {
	return views.FormPartido{
		ID:          p.IDPartido,
		RivalActual: p.Rival,
		Fecha:       p.Fecha.Format("2006-01-02"),
		TipoCancha:  strconv.Itoa(int(p.TipoCancha)),
		Rival:       p.Rival,
		Resultado:   p.Resultado,
		Posicion:    p.Posicion.String, // "" si es NULL
		Goles:       strconv.Itoa(int(p.GolesPropios)),
		Asistencias: strconv.Itoa(int(p.AsistenciasPropias)),
		Puntaje:     strconv.Itoa(int(p.PuntajePropio)),
	}
}

// aNumero convierte un texto del formulario a un entero de 32 bits (como las columnas int).
func aNumero(texto string) (int32, error) {
	n, err := strconv.ParseInt(texto, 10, 32)
	return int32(n), err
}

func noExistePartido(w http.ResponseWriter, r *http.Request) {
	mostrarPagina(w, r, http.StatusNotFound, views.PaginaError("No existe el partido", "El partido que buscás no existe o ya fue borrado."))
}

func errorInterno(w http.ResponseWriter, r *http.Request, err error) {
	log.Println("Error:", err)
	mostrarPagina(w, r, http.StatusInternalServerError, views.PaginaError("Ocurrió un error", "No se pudo completar la operación. Probá de nuevo en unos segundos."))
}
