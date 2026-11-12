package views

import (
	"context"
	"database/sql"
	"strings"
	"testing"
	"time"

	db "TuFutbol/db/sqlc"
)

// Tests de los componentes templ: se renderizan con datos de ejemplo (sin base
// de datos) y se revisa el HTML que generan.

func partido(id int32, fecha, rival, resultado string, goles, puntaje int32) db.Partido {
	f, _ := time.Parse("2006-01-02", fecha)
	return db.Partido{
		IDPartido:     id,
		Fecha:         f,
		TipoCancha:    5,
		Rival:         rival,
		Resultado:     resultado,
		Posicion:      sql.NullString{String: "Medio", Valid: true},
		GolesPropios:  goles,
		PuntajePropio: puntaje,
	}
}

// Como los devuelve ListPartidos: del más nuevo al más viejo
var ejemplo = []db.Partido{
	partido(3, "2026-10-20", "<script>alert(1)</script>", "Perdido", 0, 4),
	partido(2, "2026-10-13", "La Topadora", "Empatado", 1, 6),
	partido(1, "2026-10-06", "Los Pibes FC", "Ganado", 3, 9),
}

func html(t *testing.T, estado EstadoLista, form FormPartido, aviso string) string {
	t.Helper()
	var b strings.Builder
	if err := IndexPage("TuFutbol", ejemplo, estado, form, aviso).Render(context.Background(), &b); err != nil {
		t.Fatalf("Error al renderizar: %v", err)
	}
	return b.String()
}

func TestIndexPage(t *testing.T) {
	pagina := html(t, EstadoLista{Resaltado: 1}, FormPartido{Mensaje: "Partido guardado."}, "")

	for _, esperado := range []string{
		"<!doctype html>", // el Layout
		`<form class="formulario" method="post" action="/partidos">`,
		"vs La Topadora",
		"Hat-trick",                     // 3 goles
		`class="partido nuevo"`,         // el partido resaltado
		`<span class="nota nota-figura`, // puntaje 9
		"Partido guardado.",
		"6,3", // promedio (4 + 6 + 9) / 3
	} {
		if !strings.Contains(strings.ToLower(pagina), strings.ToLower(esperado)) {
			t.Errorf("La página no contiene %q", esperado)
		}
	}
}

func TestTextosEscapados(t *testing.T) {
	pagina := html(t, EstadoLista{}, FormPartido{}, "Borraste el partido contra <i>X</i>.")
	if strings.Contains(pagina, "<script>alert(1)</script>") || strings.Contains(pagina, "<i>X</i>") {
		t.Error("templ tendría que escapar los textos del usuario")
	}
	if !strings.Contains(pagina, "&lt;script&gt;alert(1)&lt;/script&gt;") {
		t.Error("No aparece el rival escapado")
	}
}

func TestFiltro(t *testing.T) {
	pagina := html(t, EstadoLista{Filtro: "ganado"}, FormPartido{}, "")
	if strings.Contains(pagina, "vs La Topadora") || !strings.Contains(pagina, "vs Los Pibes FC") {
		t.Error("El filtro de ganados tendría que mostrar solo los partidos ganados")
	}
	if !strings.Contains(pagina, `href="/?resultado=ganado" aria-current="true"`) {
		t.Error("El filtro elegido tendría que estar marcado con aria-current")
	}
}

func TestListaVacia(t *testing.T) {
	var b strings.Builder
	if err := PartidoList(nil, EstadoLista{}).Render(context.Background(), &b); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), "Todavía no cargaste partidos") {
		t.Error("Sin partidos tendría que mostrar el mensaje para cargar el primero")
	}
}

func TestFormularioEdicion(t *testing.T) {
	var b strings.Builder
	form := FormPartido{ID: 7, RivalActual: "Los Pibes FC", Rival: "Los Pibes FC", Puntaje: "8", Error: "El puntaje..."}
	if err := PartidoForm(form).Render(context.Background(), &b); err != nil {
		t.Fatal(err)
	}
	for _, esperado := range []string{
		"Editar partido contra Los Pibes FC",
		`action="/partidos/7/editar"`,
		"Guardar cambios",
		"nota-bueno", // el campo Puntaje con el color de su nivel
		`role="alert"`,
	} {
		if !strings.Contains(b.String(), esperado) {
			t.Errorf("El formulario de edición no contiene %q", esperado)
		}
	}
}

func TestResumenYForma(t *testing.T) {
	r := CalcularResumen(ejemplo)
	if r.Partidos != 3 || r.Goles != 4 || r.Promedio != "6,3" || r.Nivel.Clase != "regular" {
		t.Errorf("Resumen incorrecto: %+v", r)
	}
	letras := ""
	for _, item := range r.Forma {
		letras += item.Letra
	}
	if letras != "GEP" { // del más viejo al más nuevo
		t.Errorf("Se esperaba la forma GEP, se obtuvo %s", letras)
	}
}

func TestNivelesEInsignias(t *testing.T) {
	niveles := map[int32]string{1: "flojo", 4: "flojo", 5: "regular", 6: "regular", 7: "bueno", 8: "bueno", 9: "figura", 10: "figura"}
	for puntaje, clase := range niveles {
		if NivelPuntaje(puntaje).Clase != clase {
			t.Errorf("Puntaje %d: se esperaba %s", puntaje, clase)
		}
	}
	insignias := map[int32]string{2: "", 3: "Hat-trick", 4: "Póker", 5: "Manito", 7: "Manito"}
	for goles, insignia := range insignias {
		if InsigniaGoles(goles) != insignia {
			t.Errorf("%d goles: se esperaba %q", goles, insignia)
		}
	}
	resultados := map[string]string{"Ganado": "ganado", "ganamos 3-1": "ganado", "EMPATE": "empatado", "Perdido": "perdido", "Suspendido": ""}
	for resultado, clase := range resultados {
		if ClaseResultado(resultado) != clase {
			t.Errorf("Resultado %q: se esperaba %q", resultado, clase)
		}
	}
}
