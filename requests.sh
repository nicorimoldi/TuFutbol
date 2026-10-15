#!/usr/bin/env bash
#
# Pruebas de la API de TuFutbol con curl: una petición a cada endpoint,
# verificando el código HTTP de la respuesta.
#
# Uso: ./requests.sh [URL]      (por defecto http://localhost:8080)

URL="${1:-http://localhost:8080}"
FALLAS=0

# probar METODO RUTA CODIGO_ESPERADO [BODY_JSON]
probar() {
    echo "--- $1 $2 (se espera $3)"
    if [ -n "$4" ]; then
        RESPUESTA=$(curl -s -w '\n%{http_code}' -X "$1" "$URL$2" -H "Content-Type: application/json" -d "$4")
    else
        RESPUESTA=$(curl -s -w '\n%{http_code}' -X "$1" "$URL$2")
    fi
    CODIGO=$(echo "$RESPUESTA" | tail -n 1)  # última línea: el código HTTP
    CUERPO=$(echo "$RESPUESTA" | sed '$d')   # el resto: el cuerpo de la respuesta
    echo "$CUERPO"

    if [ "$CODIGO" = "$3" ]; then
        echo "OK ($CODIGO)"
    else
        echo "FALLA: se obtuvo $CODIGO"
        FALLAS=$((FALLAS + 1))
    fi
    echo
}

PARTIDO='{"fecha":"2026-10-10","tipo_cancha":5,"rival":"Los Pibes FC","resultado":"Ganado","posicion":"Delantero","goles_propios":3,"asistencias_propias":1,"puntaje_propio":9}'
PARTIDO_MODIFICADO='{"fecha":"2026-10-10","tipo_cancha":8,"rival":"Los Pibes FC","resultado":"Empatado","posicion":null,"goles_propios":1,"asistencias_propias":2,"puntaje_propio":7}'
PARTIDO_INVALIDO='{"fecha":"2026-10-10","tipo_cancha":5,"rival":"","resultado":"Ganado","goles_propios":1,"asistencias_propias":0,"puntaje_propio":8}'

# Crear
probar POST /partidos 201 "$PARTIDO"
ID=$(echo "$CUERPO" | sed -n 's/.*"id_partido":\([0-9]*\).*/\1/p')  # id del partido creado

probar POST /partidos 400 "$PARTIDO_INVALIDO"   # rival vacío

# Listar y obtener
probar GET /partidos 200
probar GET "/partidos/$ID" 200
probar GET /partidos/9999 404

# Actualizar
probar PUT "/partidos/$ID" 200 "$PARTIDO_MODIFICADO"
probar PUT /partidos/9999 404 "$PARTIDO_MODIFICADO"

# Eliminar
probar DELETE "/partidos/$ID" 204
probar GET "/partidos/$ID" 404                  # ya no existe
probar DELETE "/partidos/$ID" 404

if [ "$FALLAS" -eq 0 ]; then
    echo "Todas las pruebas pasaron"
else
    echo "Fallaron $FALLAS pruebas"
    exit 1
fi
