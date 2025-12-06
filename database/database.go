package database

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"github.com/ohhfishal/fishy/flashcard"
	_ "modernc.org/sqlite"
	"os"
)

//go:embed schema.sql
var schema string

type Store struct {
	*Queries
	db *sql.DB
}

func Connect(ctx context.Context, driver string, connection string) (*Store, error) {
	db, err := sql.Open(driver, connection)
	if err != nil {
		return nil, fmt.Errorf("opening connection: %w", err)
	}

	if err := RunMigrations(ctx, db); err != nil {
		return nil, fmt.Errorf("running migrations: %w", err)
	}
	return &Store{
		Queries: New(db),
		db:      db,
	}, nil
}

func RunMigrations(ctx context.Context, db DBTX) error {
	if _, err := db.ExecContext(ctx, schema); err != nil {
		return err
	}
	return nil
}

func (store *Store) LoadFlashcards(ctx context.Context, cards []flashcard.Flashcard) (int, error) {
	tx, err := store.db.Begin()
	if err != nil {
		return -1, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := store.WithTx(tx)
	i := 0
	for _, card := range cards {
		if _, err := qtx.InsertCard(ctx, InsertCardParams{
			Header:       card.Header,
			Description:  card.Description,
			Origin:       card.Origin,
			ClassContext: card.ClassContext,
			AiOverview:   card.AIOverview,
			Thumbnail:    card.Thumbnail,
		}); err != nil {
			return -1, err
		}
		i++
	}
	return i, tx.Commit()
}

func (store *Store) LoadFlashcardsFrom(ctx context.Context, filepath string) (int, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return -1, fmt.Errorf("opening file: %w", err)
	}
	defer file.Close()

	var flashcards []flashcard.Flashcard
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&flashcards); err != nil {
		return -1, fmt.Errorf("parsing file: %w", err)
	}
	return store.LoadFlashcards(ctx, flashcards)
}

// NOTE: This funcction must always work or there is a bug in our types
func ConvertFlashcard(oldCard Flashcard) flashcard.Flashcard {
	bytes, err := json.Marshal(oldCard)
	if err != nil {
		panic(err)
	}
	var newCard flashcard.Flashcard
	if err := json.Unmarshal(bytes, &newCard); err != nil {
		panic(err)
	}
	return newCard
}
