package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Tests de la lectura y validación del formulario. No usan la base de datos.

// formularioValido devuelve los campos de un partido válido, como los envía el navegador.
func formularioValido() url.Values {
	return url.Values{
		"fecha":               {"2026-10-10"},
		"tipo_cancha":         {"5"},
		"rival":               {"  Los Pibes FC  "},
		"resultado":           {"Ganado"},
		"posicion":            {""},
		"goles_propios":       {"3"},
		"asistencias_propias": {"1"},
		"puntaje_propio":      {"9"},
	}
}

func pedidoFormulario(campos url.Values) *http.Request {
	r := httptest.NewRequest("POST", "/partidos", strings.NewReader(campos.Encode()))
	r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return r
}

func TestLeerFormularioValido(t *testing.T) {
	form, datos, err := leerFormulario(pedidoFormulario(formularioValido()))
	if err != nil {
		t.Fatalf("Se esperaba un formulario válido, se obtuvo el error: %v", err)
	}
	if datos.Rival != "Los Pibes FC" || form.Rival != "Los Pibes FC" {
		t.Errorf("Se esperaba el rival sin espacios, se obtuvo %q", datos.Rival)
	}
	if datos.Fecha.Format("2006-01-02") != "2026-10-10" {
		t.Errorf("Fecha incorrecta: %v", datos.Fecha)
	}
	if datos.TipoCancha != 5 || datos.GolesPropios != 3 || datos.AsistenciasPropias != 1 || datos.PuntajePropio != 9 {
		t.Errorf("Números mal convertidos: %+v", datos)
	}
	if datos.Posicion.Valid {
		t.Errorf("La posición vacía se tiene que guardar como NULL")
	}
}

func TestLeerFormularioInvalido(t *testing.T) {
	casos := []struct {
		nombre, campo, valor, error string
	}{
		{"fecha vacía", "fecha", "", "La fecha"},
		{"fecha con otro formato", "fecha", "10/10/2026", "La fecha"},
		{"tipo de cancha 0", "tipo_cancha", "0", "El tipo de cancha"},
		{"tipo de cancha con letras", "tipo_cancha", "cinco", "El tipo de cancha"},
		{"rival vacío", "rival", "   ", "El rival"},
		{"rival largo", "rival", strings.Repeat("a", 51), "El rival"},
		{"resultado vacío", "resultado", "", "El resultado"},
		{"posición larga", "posicion", "Mediocampista", "La posición"},
		{"goles negativos", "goles_propios", "-1", "Los goles"},
		{"asistencias con decimales", "asistencias_propias", "1.5", "Los goles"},
		{"puntaje 0", "puntaje_propio", "0", "El puntaje"},
		{"puntaje 11", "puntaje_propio", "11", "El puntaje"},
		{"puntaje vacío", "puntaje_propio", "", "El puntaje"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			campos := formularioValido()
			campos.Set(c.campo, c.valor)
			form, _, err := leerFormulario(pedidoFormulario(campos))
			if err == nil {
				t.Fatalf("Se esperaba un error para %s=%q", c.campo, c.valor)
			}
			if !strings.HasPrefix(err.Error(), c.error) {
				t.Errorf("Mensaje inesperado: %q", err.Error())
			}
			// Lo escrito se conserva para volver a mostrarlo en el formulario
			if form.Resultado != campos.Get("resultado") {
				t.Errorf("No se conservó lo escrito en el formulario: %+v", form)
			}
		})
	}
}

func TestPosicionYLargosConTildes(t *testing.T) {
	campos := formularioValido()
	campos.Set("posicion", "Volante")
	campos.Set("rival", strings.Repeat("ñ", 50)) // 50 caracteres (100 bytes): es válido
	_, datos, err := leerFormulario(pedidoFormulario(campos))
	if err != nil {
		t.Fatalf("No se esperaba error: %v", err)
	}
	if !datos.Posicion.Valid || datos.Posicion.String != "Volante" {
		t.Errorf("Posición mal guardada: %+v", datos.Posicion)
	}
}
