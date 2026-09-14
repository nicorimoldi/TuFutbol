-- name: CreatePartido :exec
INSERT INTO partido (
    fecha, tipo_cancha, rival, resultado, 
    posicion, goles_propios, asistencias_propias, puntaje_propio
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
);

-- name: GetPartido :one
SELECT * 
FROM partido
WHERE id_partido = $1;

-- name: ListPartidos :many
SELECT * 
FROM partido
ORDER BY fecha DESC;

-- name: UpdatePartido :exec
UPDATE partido
SET 
    fecha = $2,
    tipo_cancha = $3,
    rival = $4,
    resultado = $5,
    posicion = $6,
    goles_propios = $7,
    asistencias_propias = $8,
    puntaje_propio = $9
WHERE id_partido = $1;

-- name: DeletePartido :exec
DELETE FROM partido
WHERE id_partido = $1;