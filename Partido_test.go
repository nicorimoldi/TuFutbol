package main

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	sqlc "TuFutbol/db/sqlc"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestPartidos(t *testing.T) {
	// La URL de la base la arma "make test" a partir de los valores de .env
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		t.Skip("DATABASE_URL no está definida: se omiten los tests de base de datos (usá make test)")
	}
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

	// Variable auxiliar para guardar el ID del partido creado
	var idCreado int32

	// 1. Test de CreatePartido (ahora devuelve el partido creado gracias a RETURNING)
	t.Run("Crear Partido", func(t *testing.T) {
		partido, err := queries.CreatePartido(ctx, sqlc.CreatePartidoParams{
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
		if partido.IDPartido == 0 {
			t.Fatalf("Se esperaba que la base asignara un ID al partido creado")
		}
		if partido.Rival != "Rival De Test" {
			t.Errorf("Se esperaba el rival 'Rival De Test', se obtuvo '%s'", partido.Rival)
		}

		idCreado = partido.IDPartido
	})

	// 2. Test de ListPartidos
	t.Run("Listar Partidos", func(t *testing.T) {
		partidos, err := queries.ListPartidos(ctx)
		if err != nil {
			t.Fatalf("Fallo al listar los partidos: %v", err)
		}

		encontrado := false
		for _, p := range partidos {
			if p.IDPartido == idCreado {
				encontrado = true
			}
		}
		if !encontrado {
			t.Fatalf("Se esperaba encontrar el partido con ID %d en la lista", idCreado)
		}
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

	// 4. Test de UpdatePartido (ahora devuelve el partido actualizado)
	t.Run("Actualizar Partido", func(t *testing.T) {
		partidoActualizado, err := queries.UpdatePartido(ctx, sqlc.UpdatePartidoParams{
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

		if partidoActualizado.Rival != "Rival Modificado" {
			t.Errorf("Se esperaba rival 'Rival Modificado', pero se obtuvo '%s'", partidoActualizado.Rival)
		}

		// Verificar que los cambios hayan quedado guardados en la base
		partidoGuardado, err := queries.GetPartido(ctx, idCreado)
		if err != nil {
			t.Fatalf("Fallo al obtener el partido actualizado: %v", err)
		}
		if partidoGuardado.TipoCancha != 5 {
			t.Errorf("Se esperaba tipo de cancha 5, pero se obtuvo %d", partidoGuardado.TipoCancha)
		}
	})

	// 5. Test de UpdatePartido con un ID inexistente
	t.Run("Actualizar Partido inexistente", func(t *testing.T) {
		_, err := queries.UpdatePartido(ctx, sqlc.UpdatePartidoParams{
			IDPartido:     -1,
			Fecha:         time.Now(),
			TipoCancha:    5,
			Rival:         "Nadie",
			Resultado:     "Empate",
			PuntajePropio: 5,
		})
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Se esperaba sql.ErrNoRows al actualizar un ID inexistente, se obtuvo: %v", err)
		}
	})

	// 6. Test de DeletePartido (ahora devuelve la cantidad de filas borradas)
	t.Run("Eliminar Partido", func(t *testing.T) {
		filas, err := queries.DeletePartido(ctx, idCreado)
		if err != nil {
			t.Fatalf("Fallo al eliminar el partido: %v", err)
		}
		if filas != 1 {
			t.Errorf("Se esperaba eliminar 1 fila, se eliminaron %d", filas)
		}

		// Confirmar que ya no existe en la base de datos
		_, err = queries.GetPartido(ctx, idCreado)
		if !errors.Is(err, sql.ErrNoRows) {
			t.Errorf("Se esperaba sql.ErrNoRows al buscar un registro eliminado, se obtuvo: %v", err)
		}

		// Borrar de nuevo el mismo ID no debe afectar ninguna fila
		filas, err = queries.DeletePartido(ctx, idCreado)
		if err != nil {
			t.Fatalf("Fallo al eliminar nuevamente el partido: %v", err)
		}
		if filas != 0 {
			t.Errorf("Se esperaba eliminar 0 filas, se eliminaron %d", filas)
		}
	})
}
