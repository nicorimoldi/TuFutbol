package main

import (
	"context"
	"database/sql"
	"testing"
	"time"

	sqlc "TuFutbol/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPartidos(t *testing.T) {
	connStr := "host=localhost port=5432 user=u password=p dbname=d sslmode=disable"
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		t.Fatalf("No se pudo conectar a la base de datos: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Iniciamos la transacción para todo el set de pruebas
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatalf("No se pudo iniciar la transacción: %v", err)
	}
	// 2. Al salir de TestPartidos (por éxito o por fallo), se deshace todo
	defer tx.Rollback()

	// 3. Le pasamos la transacción 'tx' a sqlc en lugar de 'db'
	queries := sqlc.New(tx)

	// 1. Test de CreatePartido
	t.Run("Crear Partido", func(t *testing.T) {
		err := queries.CreatePartido(ctx, sqlc.CreatePartidoParams{
			Fecha:              time.Now(),
			TipoCancha:         11,
			Rival:              "Rival De Test",
			Resultado:          "Ganado",
			Posicion:           sql.NullString{String: "Medio", Valid: true},
			GolesPropios:       2,
			AsistenciasPropias: 1,
			PuntajePropio:      8,
		})
		if err != nil {
			t.Fatalf("Fallo la creacion del partido: %v", err)
		}
	})

	// Variable auxiliar para capturar el ID
	var idCreado int32

	// 2. Test de ListPartidos
	t.Run("Listar Partidos", func(t *testing.T) {
		partidos, err := queries.ListPartidos(ctx)
		if err != nil {
			t.Fatalf("Fallo al listar los partidos: %v", err)
		}

		if len(partidos) == 0 {
			t.Fatalf("Se esperaba encontrar al menos un partido registrado")
		}

		// Tomamos el ID del primer partido
		idCreado = partidos[0].IDPartido
	})

	// 3. Test de GetPartido
	t.Run("Obtener Partido por ID", func(t *testing.T) {
		partido, err := queries.GetPartido(ctx, idCreado)
		if err != nil {
			t.Fatalf("Fallo al obtener el partido con ID %d: %v", idCreado, err)
		}

		if partido.Rival != "Rival De Test" {
			t.Errorf("Se esperaba el rival 'Rival De Test', se obtuvo '%s'", partido.Rival)
		}
	})

	// 4. Test de UpdatePartido
	t.Run("Actualizar Partido", func(t *testing.T) {
		err := queries.UpdatePartido(ctx, sqlc.UpdatePartidoParams{
			IDPartido:          idCreado,
			Fecha:              time.Now(),
			TipoCancha:         5,
			Rival:              "Rival Modificado",
			Resultado:          "Empate",
			Posicion:           sql.NullString{String: "Defensa", Valid: true},
			GolesPropios:       0,
			AsistenciasPropias: 0,
			PuntajePropio:      6,
		})
		if err != nil {
			t.Fatalf("Fallo la actualizacion del partido: %v", err)
		}

		// Verificar que se hayan aplicado los cambios
		partidoActualizado, err := queries.GetPartido(ctx, idCreado)
		if err != nil {
			t.Fatalf("Fallo al obtener el partido actualizado: %v", err)
		}

		if partidoActualizado.Rival != "Rival Modificado" {
			t.Errorf("Se esperaba rival 'Rival Modificado', pero se obtuvo '%s'", partidoActualizado.Rival)
		}
	})

	// 5. Test de DeletePartido
	t.Run("Eliminar Partido", func(t *testing.T) {
		err := queries.DeletePartido(ctx, idCreado)
		if err != nil {
			t.Fatalf("Fallo al eliminar el partido: %v", err)
		}

		// Confirmar que ya no existe en la base de datos
		_, err = queries.GetPartido(ctx, idCreado)
		if err == nil {
			t.Errorf("Se esperaba un error al buscar un registro eliminado, pero no se produjo ninguno")
		}
	})
}
