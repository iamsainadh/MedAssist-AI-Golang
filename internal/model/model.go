package model

import "mime/multipart"

type UploadRequest struct {
	Document *multipart.FileHeader `form:"document" binding:"required"`
}

type Chunk struct {
	DocumentId string
	ChunkIndex int // Considering Page Index Because of no sections
	Content    string
	Embedding  []float32
}

type Document struct {
	Filename string
	FileHash string
}
