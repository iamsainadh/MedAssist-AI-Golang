package repository

import (
	"context"
	"fmt"
	"iamsainadh/medassist_ai/internal/model"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type Repository interface {
	CreateDocument(ctx context.Context, document model.Document) (string, error)
	CreateChunk(ctx context.Context, chunks []model.Chunk) error
	SearchChunk(ctx context.Context, documentId string, queryEmbedding []float32, topK int) (string, error)
	GetDocumentByHash(ctx context.Context, hash string) (string, error)
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		Db: pool,
	}
}

type repository struct {
	Db *pgxpool.Pool
}

func (r *repository) CreateDocument(ctx context.Context, document model.Document) (string, error) {
	// currCtx, cancel := context.WithTimeout(ctx, time.Second*10)

	// defer cancel()

	query := `INSERT into documents(filename,file_hash) values($1,$2) returning id`

	var document_id string

	err := r.Db.QueryRow(
		ctx,
		query,
		document.Filename,
		document.FileHash,
	).Scan(
		&document_id,
	)

	if err != nil {
		return "", err
	}

	return document_id, nil

}

// ! Insert Chunk
func (r *repository) CreateChunk(ctx context.Context, chunks []model.Chunk) error {
	// currCtx, cancel := context.WithTimeout(ctx, time.Second*10)
	// defer cancel()

	tx, err := r.Db.Begin(ctx)

	if err != nil {
		return err
	}

	defer func() {
		fmt.Println("Error Here")
		tx.Rollback(ctx)
	}()

	batch := &pgx.Batch{}

	query := `INSERT into chunks(document_id,chunk_index,content,embedding) values($1,$2,$3,$4)`

	for _, c := range chunks {
		batch.Queue(
			query,
			c.DocumentId,
			c.ChunkIndex,
			c.Content,
			pgvector.NewVector(c.Embedding),
		)
	}

	br := tx.SendBatch(ctx, batch)

	for range chunks {
		if _, err := br.Exec(); err != nil {
			return err
		}
	}

	if err := br.Close(); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *repository) SearchChunk(ctx context.Context, documentId string, queryEmbedding []float32, topK int) (string, error) {

	rows, err := r.Db.Query(ctx,
		`SELECT content
		FROM chunks
		WHERE document_id = $1
		ORDER BY embedding <=> $2
		LIMIT $3`,
		documentId, pgvector.NewVector(queryEmbedding), topK)

	if err != nil {
		return "", err
	}
	defer rows.Close()

	var chunks []string
	for rows.Next() {
		var content string

		if err := rows.Scan(&content); err != nil {
			return "", err
		}

		chunks = append(chunks, content)
	}

	return strings.Join(chunks, "\n\n"), rows.Err()
}

func (r *repository) GetDocumentByHash(ctx context.Context, hash string) (string, error) {
	var id string

	err := r.Db.QueryRow(
		ctx,
		`SELECT id FROM documents WHERE file_hash = $1`,
		hash,
	).Scan(&id)

	return id, err
}
