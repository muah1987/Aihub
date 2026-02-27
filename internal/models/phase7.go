package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// KnowledgeDocument represents an uploaded document for knowledge extraction.
type KnowledgeDocument struct {
	ID           uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID    uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	UploadedBy   *uuid.UUID     `gorm:"type:uuid" json:"uploaded_by,omitempty"`
	Title        string         `gorm:"size:500;not null" json:"title"`
	FileName     string         `gorm:"size:500;not null" json:"file_name"`
	FileType     string         `gorm:"size:50;not null" json:"file_type"`
	FileSize     int64          `gorm:"not null;default:0" json:"file_size"`
	ContentHash  string         `gorm:"size:64" json:"content_hash,omitempty"`
	ChunkCount   int            `gorm:"not null;default:0" json:"chunk_count"`
	Status       string         `gorm:"size:20;not null;default:pending" json:"status"`
	ErrorMessage string         `gorm:"type:text" json:"error_message,omitempty"`
	Metadata     datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	Chunks       []DocumentChunk `gorm:"foreignKey:DocumentID" json:"chunks,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

func (d *KnowledgeDocument) BeforeCreate(tx *gorm.DB) error {
	if d.ID == uuid.Nil {
		d.ID = uuid.New()
	}
	return nil
}

// DocumentChunk is a segment of a document used for retrieval.
type DocumentChunk struct {
	ID          uuid.UUID      `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	DocumentID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"document_id"`
	ProjectID   uuid.UUID      `gorm:"type:uuid;not null;index" json:"project_id"`
	ChunkIndex  int            `gorm:"not null" json:"chunk_index"`
	Content     string         `gorm:"type:text;not null" json:"content"`
	TokenCount  int            `gorm:"not null;default:0" json:"token_count"`
	Metadata    datatypes.JSON `gorm:"type:jsonb;default:'{}'" json:"metadata"`
	CreatedAt   time.Time      `json:"created_at"`
}

func (c *DocumentChunk) BeforeCreate(tx *gorm.DB) error {
	if c.ID == uuid.Nil {
		c.ID = uuid.New()
	}
	return nil
}

// MemoryVersion tracks changes to a TeamMemory entry.
type MemoryVersion struct {
	ID            uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	MemoryID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"memory_id"`
	VersionNumber int        `gorm:"not null" json:"version_number"`
	Content       string     `gorm:"type:text;not null" json:"content"`
	ContentType   string     `gorm:"size:50;default:text" json:"content_type"`
	ChangedBy     *uuid.UUID `gorm:"type:uuid" json:"changed_by,omitempty"`
	ChangeType    string     `gorm:"size:20;not null;default:update" json:"change_type"`
	DiffSummary   string     `gorm:"type:text" json:"diff_summary,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (v *MemoryVersion) BeforeCreate(tx *gorm.DB) error {
	if v.ID == uuid.Nil {
		v.ID = uuid.New()
	}
	return nil
}

// SharedMemory enables cross-project knowledge sharing within an organization.
type SharedMemory struct {
	ID              uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	OrganizationID  uuid.UUID  `gorm:"type:uuid;not null;index" json:"organization_id"`
	SourceProjectID uuid.UUID  `gorm:"type:uuid;not null;index" json:"source_project_id"`
	SourceMemoryID  uuid.UUID  `gorm:"type:uuid;not null" json:"source_memory_id"`
	SharedBy        *uuid.UUID `gorm:"type:uuid" json:"shared_by,omitempty"`
	AccessLevel     string     `gorm:"size:20;not null;default:read" json:"access_level"`
	CreatedAt       time.Time  `json:"created_at"`

	Memory  TeamMemory `gorm:"foreignKey:SourceMemoryID" json:"memory,omitempty"`
	Project Project    `gorm:"foreignKey:SourceProjectID" json:"project,omitempty"`
}

func (s *SharedMemory) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

// AutoExtraction stores automatically extracted knowledge from chat messages.
type AutoExtraction struct {
	ID               uuid.UUID  `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	ProjectID        uuid.UUID  `gorm:"type:uuid;not null;index" json:"project_id"`
	MessageID        *uuid.UUID `gorm:"type:uuid" json:"message_id,omitempty"`
	ExtractedContent string     `gorm:"type:text;not null" json:"extracted_content"`
	ExtractionType   string     `gorm:"size:50;not null" json:"extraction_type"`
	Confidence       float64    `gorm:"type:decimal(3,2);not null;default:0.5" json:"confidence"`
	Accepted         *bool      `gorm:"type:boolean" json:"accepted"`
	MemoryID         *uuid.UUID `gorm:"type:uuid" json:"memory_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

func (a *AutoExtraction) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}
