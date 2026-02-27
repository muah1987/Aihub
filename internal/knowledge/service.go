package knowledge

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var (
	ErrDocumentNotFound = errors.New("document not found")
	ErrChunkNotFound    = errors.New("chunk not found")
)

const (
	DefaultChunkSize    = 1000 // tokens (approximate by words)
	DefaultChunkOverlap = 100
)

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

// ---- Document CRUD ----

type CreateDocumentInput struct {
	Title    string `json:"title"`
	FileName string `json:"file_name"`
	FileType string `json:"file_type"`
	Content  string `json:"content"` // raw text content
}

func (s *Service) CreateDocument(projectID uuid.UUID, userID *uuid.UUID, input *CreateDocumentInput) (*models.KnowledgeDocument, error) {
	hash := sha256.Sum256([]byte(input.Content))
	hashStr := hex.EncodeToString(hash[:])

	doc := &models.KnowledgeDocument{
		ProjectID:   projectID,
		UploadedBy:  userID,
		Title:       input.Title,
		FileName:    input.FileName,
		FileType:    input.FileType,
		FileSize:    int64(len(input.Content)),
		ContentHash: hashStr,
		Status:      "processing",
	}

	if err := s.db.Create(doc).Error; err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	// Chunk the content
	chunks := chunkText(input.Content, DefaultChunkSize, DefaultChunkOverlap)
	var dbChunks []models.DocumentChunk
	for i, chunk := range chunks {
		dbChunks = append(dbChunks, models.DocumentChunk{
			DocumentID: doc.ID,
			ProjectID:  projectID,
			ChunkIndex: i,
			Content:    chunk,
			TokenCount: estimateTokens(chunk),
		})
	}

	if len(dbChunks) > 0 {
		if err := s.db.CreateInBatches(dbChunks, 100).Error; err != nil {
			s.db.Model(doc).Updates(map[string]interface{}{
				"status":        "error",
				"error_message": fmt.Sprintf("chunking failed: %v", err),
			})
			return doc, fmt.Errorf("failed to create chunks: %w", err)
		}
	}

	doc.ChunkCount = len(dbChunks)
	doc.Status = "ready"
	s.db.Model(doc).Updates(map[string]interface{}{
		"status":      "ready",
		"chunk_count": len(dbChunks),
	})

	return doc, nil
}

func (s *Service) GetDocument(projectID, docID uuid.UUID) (*models.KnowledgeDocument, error) {
	var doc models.KnowledgeDocument
	err := s.db.Where("id = ? AND project_id = ?", docID, projectID).First(&doc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrDocumentNotFound
		}
		return nil, err
	}
	return &doc, nil
}

func (s *Service) ListDocuments(projectID uuid.UUID) ([]models.KnowledgeDocument, error) {
	var docs []models.KnowledgeDocument
	err := s.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&docs).Error
	return docs, err
}

func (s *Service) DeleteDocument(projectID, docID uuid.UUID) error {
	result := s.db.Where("id = ? AND project_id = ?", docID, projectID).Delete(&models.KnowledgeDocument{})
	if result.RowsAffected == 0 {
		return ErrDocumentNotFound
	}
	return result.Error
}

// ---- Chunk retrieval ----

func (s *Service) GetChunks(projectID, docID uuid.UUID) ([]models.DocumentChunk, error) {
	var chunks []models.DocumentChunk
	err := s.db.Where("document_id = ? AND project_id = ?", docID, projectID).
		Order("chunk_index ASC").Find(&chunks).Error
	return chunks, err
}

func (s *Service) SearchChunks(projectID uuid.UUID, query string, limit int) ([]models.DocumentChunk, error) {
	if limit <= 0 {
		limit = 20
	}
	var chunks []models.DocumentChunk
	err := s.db.Where("project_id = ? AND to_tsvector('english', content) @@ plainto_tsquery('english', ?)", projectID, query).
		Order("ts_rank(to_tsvector('english', content), plainto_tsquery('english', '" + sanitizeQuery(query) + "')) DESC").
		Limit(limit).
		Find(&chunks).Error
	return chunks, err
}

// ---- Helpers ----

func chunkText(text string, chunkSize, overlap int) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	if len(words) <= chunkSize {
		return []string{text}
	}

	var chunks []string
	start := 0
	for start < len(words) {
		end := start + chunkSize
		if end > len(words) {
			end = len(words)
		}
		chunk := strings.Join(words[start:end], " ")
		chunks = append(chunks, chunk)
		start += chunkSize - overlap
		if start >= len(words) {
			break
		}
	}
	return chunks
}

func estimateTokens(text string) int {
	// Rough estimate: ~0.75 tokens per word for English
	words := len(strings.Fields(text))
	return int(float64(words) * 1.33)
}

func sanitizeQuery(q string) string {
	// Remove single quotes to prevent SQL issues in ts_rank
	return strings.ReplaceAll(q, "'", "")
}
