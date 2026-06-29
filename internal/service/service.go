package service

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"iamsainadh/medassist_ai/internal/model"
	"iamsainadh/medassist_ai/internal/ollama"
	"iamsainadh/medassist_ai/internal/repository"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"unicode"

	"github.com/ledongthuc/pdf"
)

// ! PDFType
type PdfPage struct {
	PageNumber int
	Content    []string
}

type Service struct {
	ollamaClient *ollama.Client
	repo         repository.Repository
}

func NewService(ollamaClient *ollama.Client, repo repository.Repository) *Service {
	return &Service{
		ollamaClient: ollamaClient,
		repo:         repo,
	}
}
func (s *Service) UploadDocument(ctx context.Context, req *model.UploadRequest) (string, error) {
	var pages []PdfPage
	file, err := req.Document.Open()

	if err != nil {
		log.Fatal(err)
	}

	defer file.Close()

	fileHash, err := ComputeHash(file)

	existingDocId, err := s.repo.GetDocumentByHash(ctx, fileHash)

	if err == nil {
		return existingDocId, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return "", err
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	pages, err = ExtractText(file, pages)

	if err != nil {
		log.Fatal(err)
	}

	pages = RemoveRepeatedMetaData(pages)

	documentId, err := s.repo.CreateDocument(ctx, model.Document{
		Filename: req.Document.Filename,
		FileHash: fileHash,
	})

	chunks, err := s.CreateChunk(documentId, pages)

	err = s.repo.CreateChunk(ctx, chunks)

	if err != nil {
		return "", err
	}
	return documentId, nil
}

func ExtractText(fileReader io.Reader, pages []PdfPage) ([]PdfPage, error) {
	data, err := io.ReadAll(fileReader)

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	reader := bytes.NewReader(data)

	pdfReader, err := pdf.NewReader(reader, reader.Size())

	if err != nil {
		return nil, err
	}

	for pageIndex := 1; pageIndex <= pdfReader.NumPage(); pageIndex++ {
		page := pdfReader.Page(pageIndex)

		if page.V.IsNull() {
			continue
		}

		extractedText, err := page.GetPlainText(nil)

		if err != nil {
			return nil, err
		}

		//! Normalize The Text Remove the Extra Space
		lines := normalizeText(extractedText)

		pages = append(pages, PdfPage{
			PageNumber: pageIndex,
			Content:    lines,
		})
	}

	return pages, nil
}

func normalizeText(extractedText string) []string {
	raw := strings.Split(extractedText, "\n")

	lines := make([]string, 0, len(raw))

	for _, line := range raw {
		if line == "" {
			continue
		}

		if !containsAlphaNum(line) {
			continue
		}

		lines = append(lines, line)
	}

	return lines
}

func containsAlphaNum(text string) bool {
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return true
		}
	}

	return false
}

func RemoveRepeatedMetaData(pages []PdfPage) []PdfPage {
	contentLineCount := make(map[string]int)

	for _, page := range pages {
		seen := make(map[string]bool)

		for _, content := range page.Content {
			content = strings.TrimSpace(content)

			if content == "" {
				continue
			}

			if seen[content] {
				continue
			}

			seen[content] = true

			contentLineCount[content]++
		}
	}

	ignore := make(map[string]bool)
	totalPages := len(pages)

	for line, count := range contentLineCount {
		if count >= totalPages-1 {

			ignore[line] = true
		}
	}

	for i := range pages {

		var cleaned []string

		for _, line := range pages[i].Content {

			line = strings.TrimSpace(line)

			if ignore[line] {
				continue
			}

			cleaned = append(cleaned, line)
		}

		pages[i].Content = cleaned
	}

	return pages
}

func (s *Service) CreateChunk(documentId string, pages []PdfPage) ([]model.Chunk, error) {
	chunks := make([]model.Chunk, 0, len(pages))

	for i, page := range pages {

		text := strings.Join(page.Content, "\n")

		embedding, err := s.ollamaClient.Embed(text)
		if err != nil {
			return nil, err
		}

		chunks = append(chunks, model.Chunk{
			DocumentId: documentId,
			ChunkIndex: i,
			Content:    text,
			Embedding:  embedding,
		})
	}

	return chunks, nil
}

func ComputeHash(file multipart.File) (string, error) {
	hash := sha256.New()

	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

//! Add Single Page Chunk to split chunks
// func (s *Service) CreateChunk(documentId string, pages []PdfPage) ([]model.Chunk, error) {
// 	total := 0
// 	for _, p := range pages {
// 		total += len(p.Content)
// 	}

// 	chunks := make([]model.Chunk, 0, total)

// 	chunkIndex := 0

// 	for _, page := range pages {
// 		embeddings, err := s.ollamaClient.EmbedBatch(page.Content)
// 		if err != nil {
// 			return nil, err
// 		}

// 		for i, text := range page.Content {
// 			chunks = append(chunks, model.Chunk{
// 				DocumentId: documentId,
// 				ChunkIndex: chunkIndex,
// 				Content:    text,
// 				Embedding:  embeddings[i],
// 			})
// 			chunkIndex++
// 		}
// 	}

// 	return chunks, nil
// }
