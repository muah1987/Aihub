package terminal

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/muah1987/Aihub/internal/models"
	"gorm.io/gorm"
)

var (
	ErrSessionNotFound = errors.New("terminal session not found")
)

type ActiveSession struct {
	Session   *models.TerminalSession
	Cancel    context.CancelFunc
	CreatedAt time.Time
}

type Service struct {
	db        *gorm.DB
	container *ContainerManager
	sessions  map[uuid.UUID]*ActiveSession
	mu        sync.RWMutex
}

func NewService(db *gorm.DB, container *ContainerManager) *Service {
	return &Service{
		db:        db,
		container: container,
		sessions:  make(map[uuid.UUID]*ActiveSession),
	}
}

func (s *Service) CreateSession(ctx context.Context, projectID, userID uuid.UUID, projectName string) (*models.TerminalSession, error) {
	containerCtx, cancel := context.WithCancel(context.Background())

	info, err := s.container.CreateContainer(containerCtx, projectName)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to create sandbox container: %w", err)
	}

	session := &models.TerminalSession{
		ProjectID:    projectID,
		UserID:       userID,
		ContainerID:  info.ID,
		Status:       "running",
		LastActivity: time.Now(),
	}

	if err := s.db.Create(session).Error; err != nil {
		cancel()
		s.container.StopContainer(ctx, info.ID)
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	s.mu.Lock()
	s.sessions[session.ID] = &ActiveSession{
		Session:   session,
		Cancel:    cancel,
		CreatedAt: time.Now(),
	}
	s.mu.Unlock()

	return session, nil
}

func (s *Service) GetSession(sessionID uuid.UUID) (*models.TerminalSession, error) {
	s.mu.RLock()
	active, ok := s.sessions[sessionID]
	s.mu.RUnlock()
	if ok {
		return active.Session, nil
	}

	var session models.TerminalSession
	if err := s.db.First(&session, "id = ?", sessionID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to get session: %w", err)
	}
	return &session, nil
}

func (s *Service) ListSessions(projectID uuid.UUID) ([]models.TerminalSession, error) {
	var sessions []models.TerminalSession
	if err := s.db.Where("project_id = ? AND status != ?", projectID, "stopped").
		Order("created_at DESC").Find(&sessions).Error; err != nil {
		return nil, fmt.Errorf("failed to list sessions: %w", err)
	}
	return sessions, nil
}

func (s *Service) StopSession(ctx context.Context, sessionID uuid.UUID) error {
	session, err := s.GetSession(sessionID)
	if err != nil {
		return err
	}

	if session.ContainerID != "" {
		s.container.StopContainer(ctx, session.ContainerID)
	}

	s.db.Model(&models.TerminalSession{}).Where("id = ?", sessionID).Update("status", "stopped")

	s.mu.Lock()
	if active, ok := s.sessions[sessionID]; ok {
		active.Cancel()
		delete(s.sessions, sessionID)
	}
	s.mu.Unlock()

	return nil
}

func (s *Service) GetContainerManager() *ContainerManager {
	return s.container
}

func (s *Service) UpdateActivity(sessionID uuid.UUID) {
	s.db.Model(&models.TerminalSession{}).Where("id = ?", sessionID).Update("last_activity", time.Now())
}
