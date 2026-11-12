#!/usr/bin/env bash
#
# Pruebas de la aplicación web de TuFutbol con curl: pide cada página y envía
# los formularios como lo haría el navegador, verificando el código HTTP, la
# redirección (patrón Post/Redirect/Get) y que el HTML tenga lo esperado.
#
# Uso: ./requests.sh [URL]      (por defecto http://localhost:8080)

URL="${1:-http://localhost:8080}"
FALLAS=0
CUERPO=""
DESTINO=""

# probar METODO RUTA CODIGO_ESPERADO [CAMPO=VALOR ...]
# Los campos se envían como un formulario HTML (application/x-www-form-urlencoded).
# No sigue las redirecciones: así se puede ver el 303 y a dónde redirige.
probar() {
    local metodo="$1" ruta="$2" esperado="$3"
    shift 3
    local datos=()
    for campo in "$@"; do
        datos+=(--data-urlencode "$campo")
    done

    echo "--- $metodo $ruta (se espera $esperado)"
    local respuesta
    respuesta=$(curl -s -X "$metodo" "${datos[@]}" -w '\n%{http_code} %{redirect_url}' "$URL$ruta")
    local ultima
    ultima=$(echo "$respuesta" | tail -n 1)  # última línea: código HTTP y destino de la redirección
    CUERPO=$(echo "$respuesta" | sed '$d')   # el resto: el HTML de la respuesta
    local codigo="${ultima%% *}"
    DESTINO="${ultima#* }"

    if [ "$codigo" = "$esperado" ]; then
        echo "OK ($codigo${DESTINO:+ -> $DESTINO})"
    else
        echo "FALLA: se obtuvo $codigo"
        FALLAS=$((FALLAS + 1))
    fi
}

# contiene TEXTO: verifica que el HTML de la última respuesta tenga el texto
contiene() {
    if grep -qF -- "$1" <<< "$CUERPO"; then
        echo "    OK: la página contiene '$1'"
    else
        echo "    FALLA: la página no contiene '$1'"
        FALLAS=$((FALLAS + 1))
    fi
}

# no_contiene TEXTO: verifica que el HTML de la última respuesta NO tenga el texto
no_contiene() {
    if grep -qF -- "$1" <<< "$CUERPO"; then
        echo "    FALLA: la página contiene '$1'"
        FALLAS=$((FALLAS + 1))
    else
        echo "    OK: la página no contiene '$1'"
    fi
}

PARTIDO=(fecha=2026-10-10 tipo_cancha=5 "rival=Los Pibes FC" resultado=Ganado posicion=Delantero
    goles_propios=3 asistencias_propias=1 puntaje_propio=9)
PARTIDO_MODIFICADO=(fecha=2026-10-10 tipo_cancha=8 "rival=Los Pibes FC" resultado=Empatado posicion=
    goles_propios=1 asistencias_propias=2 puntaje_propio=7)
PARTIDO_INVALIDO=(fecha=2026-10-10 tipo_cancha=5 rival= resultado=Ganado
    goles_propios=1 asistencias_propias=0 puntaje_propio=8)

# Página principal sin partidos
probar GET / 200
contiene "Todavía no cargaste partidos"
contiene 'action="/partidos"'

# Crear: el formulario redirige (303) a la página principal
probar POST /partidos 303 "${PARTIDO[@]}"
ID=$(echo "$DESTINO" | sed -n 's/.*guardado=\([0-9]*\).*/\1/p')  # id del partido creado
probar GET "/?guardado=$ID" 200
contiene "Partido contra Los Pibes FC guardado."
contiene "vs Los Pibes FC"
contiene "Hat-trick"

# Datos inválidos: se vuelve a mostrar el formulario con el error y lo escrito (422)
probar POST /partidos 422 "${PARTIDO_INVALIDO[@]}"
contiene "El rival es obligatorio"
contiene 'value="2026-10-10"'

# Los textos se escapan: un rival con HTML no se interpreta como HTML
probar POST /partidos 303 fecha=2026-10-11 tipo_cancha=5 "rival=<b>Malicioso</b>" resultado=Perdido \
    goles_propios=0 asistencias_propias=0 puntaje_propio=3
ID_MALICIOSO=$(echo "$DESTINO" | sed -n 's/.*guardado=\([0-9]*\).*/\1/p')
probar GET / 200
contiene "vs &lt;b&gt;Malicioso&lt;/b&gt;"
no_contiene "<b>Malicioso</b>"

# Filtros
probar GET "/?resultado=perdido" 200
contiene "Malicioso"
no_contiene "vs Los Pibes FC"

# Editar
probar GET "/partidos/$ID/editar" 200
contiene "Editar partido contra Los Pibes FC"
contiene 'value="Delantero"'
probar GET /partidos/9999/editar 404
probar POST "/partidos/$ID/editar" 303 "${PARTIDO_MODIFICADO[@]}"
probar GET "/?editado=$ID" 200
contiene "Cambios guardados en el partido contra Los Pibes FC."
contiene "Fútbol 8"
probar POST /partidos/9999/editar 404 "${PARTIDO_MODIFICADO[@]}"

# Borrar
probar POST "/partidos/$ID/borrar" 303
probar GET "/?borrado=Los+Pibes+FC" 200
contiene "Borraste el partido contra Los Pibes FC."
no_contiene "vs Los Pibes FC"
probar POST "/partidos/$ID/borrar" 404     # ya no existe
probar POST "/partidos/$ID_MALICIOSO/borrar" 303

# Archivos estáticos y rutas que no existen
probar GET /static/styles.css 200
probar GET /no-existe 404

echo
if [ "$FALLAS" -eq 0 ]; then
    echo "Todas las pruebas pasaron"
else
    echo "Fallaron $FALLAS pruebas"
    exit 1
fi
