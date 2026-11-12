package views

import (
	"math"
	"strconv"
	"strings"

	db "TuFutbol/db/sqlc"

	"github.com/a-h/templ"
)

// Este archivo tiene los datos que reciben los componentes templ y las funciones
// que usan para mostrar los partidos (antes estaban en app.js, en el navegador).

// FormPartido son los valores del formulario de un partido, tal como se escriben
// en los inputs (por eso son todos texto). Si el servidor rechaza los datos,
// el formulario se vuelve a mostrar con lo que se había escrito y el error.
type FormPartido struct {
	ID          int32  // 0: se está cargando un partido nuevo; otro valor: se edita ese partido
	RivalActual string // rival guardado del partido que se edita (para el título)

	Fecha       string
	TipoCancha  string
	Rival       string
	Resultado   string
	Posicion    string
	Goles       string
	Asistencias string
	Puntaje     string

	Error   string // mensaje de error (datos inválidos)
	Mensaje string // mensaje de éxito (partido guardado)
}

// EstadoLista indica cómo mostrar la lista de partidos. Llega en la URL
// (ej. /?resultado=ganado&guardado=12), así se conserva al recargar la página.
type EstadoLista struct {
	Filtro    string // "" (todos), "ganado", "empatado" o "perdido"
	Resaltado int32  // id del partido recién guardado (0: ninguno)
	Editando  int32  // id del partido que se está editando (0: ninguno)
}

// Nivel es la clasificación del puntaje, que define su color.
type Nivel struct {
	Clase  string // sufijo de la clase de CSS: nota-flojo, nota-regular...
	Nombre string // lo que se muestra y leen los lectores de pantalla
}

// Resumen son los totales que se muestran arriba de la lista.
type Resumen struct {
	Partidos    int
	Goles       int32
	Asistencias int32
	Promedio    string // con coma decimal: "7,5"
	Nivel       Nivel
	Forma       []ItemForma // últimos 5 resultados, del más viejo al más nuevo
}

// ItemForma es uno de los últimos 5 resultados (G / E / P).
type ItemForma struct {
	Clase       string
	Letra       string
	Descripcion string
}

// Filtro es un botón de filtro de la lista.
type Filtro struct {
	Valor  string // valor del parámetro ?resultado= ("" para todos)
	Nombre string
	Cuenta int
}

var meses = []string{"ene", "feb", "mar", "abr", "may", "jun", "jul", "ago", "sep", "oct", "nov", "dic"}

var nombresFiltro = map[string]string{"ganado": "ganados", "empatado": "empatados", "perdido": "perdidos"}

// FiltroValido devuelve el filtro si es uno de los conocidos, o "" (todos).
func FiltroValido(filtro string) string {
	if _, ok := nombresFiltro[filtro]; ok {
		return filtro
	}
	return ""
}

// NivelPuntaje clasifica el puntaje (1 a 10) para darle color.
func NivelPuntaje(puntaje int32) Nivel {
	switch {
	case puntaje >= 9:
		return Nivel{"figura", "Figura"}
	case puntaje >= 7:
		return Nivel{"bueno", "Bueno"}
	case puntaje >= 5:
		return Nivel{"regular", "Regular"}
	default:
		return Nivel{"flojo", "Flojo"}
	}
}

// ClaseResultado devuelve una clase de CSS según el resultado, para darle color.
// El resultado es texto libre: si no es ganado/empatado/perdido, queda sin color.
func ClaseResultado(resultado string) string {
	texto := strings.ToLower(strings.TrimSpace(resultado))
	switch {
	case strings.HasPrefix(texto, "gan"):
		return "ganado"
	case strings.HasPrefix(texto, "emp"):
		return "empatado"
	case strings.HasPrefix(texto, "perd"):
		return "perdido"
	}
	return ""
}

// InsigniaGoles devuelve la insignia según los goles del partido ("" si no corresponde).
func InsigniaGoles(goles int32) string {
	switch {
	case goles >= 5:
		return "Manito"
	case goles == 4:
		return "Póker"
	case goles == 3:
		return "Hat-trick"
	}
	return ""
}

// CalcularResumen suma los totales de todos los partidos. La lista viene
// ordenada del partido más nuevo al más viejo (ListPartidos).
func CalcularResumen(partidos []db.Partido) Resumen {
	r := Resumen{Partidos: len(partidos)}
	if len(partidos) == 0 {
		return r
	}
	var puntajes int32
	for _, p := range partidos {
		r.Goles += p.GolesPropios
		r.Asistencias += p.AsistenciasPropias
		puntajes += p.PuntajePropio
	}
	promedio := float64(puntajes) / float64(len(partidos))
	r.Nivel = NivelPuntaje(int32(math.Round(promedio)))
	r.Promedio = strings.Replace(strconv.FormatFloat(math.Round(promedio*10)/10, 'f', -1, 64), ".", ",", 1)

	// Últimos 5, del más viejo al más nuevo (el de la derecha es el último partido)
	ultimos := partidos[:min(5, len(partidos))]
	letras := map[string]string{"ganado": "G", "empatado": "E", "perdido": "P"}
	for i := len(ultimos) - 1; i >= 0; i-- {
		p := ultimos[i]
		clase := ClaseResultado(p.Resultado)
		letra, ok := letras[clase]
		if !ok {
			letra = "–"
		}
		r.Forma = append(r.Forma, ItemForma{
			Clase:       clase,
			Letra:       letra,
			Descripcion: p.Resultado + " contra " + p.Rival,
		})
	}
	return r
}

// Filtros devuelve los botones de filtro con la cantidad de partidos de cada uno.
func Filtros(partidos []db.Partido) []Filtro {
	cuentas := map[string]int{}
	for _, p := range partidos {
		cuentas[ClaseResultado(p.Resultado)]++
	}
	return []Filtro{
		{"", "Todos", len(partidos)},
		{"ganado", "Ganados", cuentas["ganado"]},
		{"empatado", "Empatados", cuentas["empatado"]},
		{"perdido", "Perdidos", cuentas["perdido"]},
	}
}

// Filtrar devuelve los partidos del filtro elegido ("" devuelve todos).
func Filtrar(partidos []db.Partido, filtro string) []db.Partido {
	if filtro == "" {
		return partidos
	}
	var visibles []db.Partido
	for _, p := range partidos {
		if ClaseResultado(p.Resultado) == filtro {
			visibles = append(visibles, p)
		}
	}
	return visibles
}

// Funciones cortas que usan los componentes templ

func urlFiltro(filtro string) string {
	if filtro == "" {
		return "/"
	}
	return "/?resultado=" + filtro
}

// filtroElegido marca con aria-current="true" el filtro que se está mostrando.
func filtroElegido(elegido bool) templ.Attributes {
	if elegido {
		return templ.Attributes{"aria-current": "true"}
	}
	return templ.Attributes{}
}

// dia y mesAnio arman la fecha de la lista: "5" y "oct 2026".
func dia(p db.Partido) string {
	return strconv.Itoa(p.Fecha.Day())
}

func mesAnio(p db.Partido) string {
	return meses[p.Fecha.Month()-1] + " " + strconv.Itoa(p.Fecha.Year())
}

func detalle(p db.Partido) string {
	texto := "Fútbol " + strconv.Itoa(int(p.TipoCancha))
	if p.Posicion.Valid && p.Posicion.String != "" {
		texto += ", " + p.Posicion.String
	}
	return texto
}

func numero(n int32) string { return strconv.Itoa(int(n)) }

// clasePuntajeFormulario pinta el campo Puntaje del formulario con el color de
// su nivel cuando ya tiene un valor válido (al editar o si hubo un error).
func clasePuntajeFormulario(valor string) string {
	puntaje, err := strconv.Atoi(valor)
	if err != nil || puntaje < 1 || puntaje > 10 {
		return ""
	}
	return "nota-" + NivelPuntaje(int32(puntaje)).Clase
}

// claseItem arma las clases del <li> de un partido.
func claseItem(p db.Partido, estado EstadoLista) string {
	clase := "partido"
	if p.IDPartido == estado.Resaltado {
		clase += " nuevo"
	}
	if p.IDPartido == estado.Editando {
		clase += " editando"
	}
	return clase
}
