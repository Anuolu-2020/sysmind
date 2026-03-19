package services

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"sysmind/internal/models"
)

// ChatService manages chat sessions and persistence
type ChatService struct {
	dataDir  string
	sessions map[string]*models.ChatSession
	mu       sync.RWMutex
}

// NewChatService creates a new chat service
func NewChatService() (*ChatService, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	dataDir := filepath.Join(configDir, "sysmind", "chats")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return nil, err
	}

	cs := &ChatService{
		dataDir:  dataDir,
		sessions: make(map[string]*models.ChatSession),
	}

	// Load existing sessions
	cs.loadAllSessions()

	return cs, nil
}

// generateID creates a unique session/message ID
func generateID() string {
	bytes := make([]byte, 8)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// nowMs returns current time in milliseconds
func nowMs() int64 {
	return time.Now().UnixMilli()
}

// loadAllSessions loads all chat sessions from disk
func (cs *ChatService) loadAllSessions() {
	files, err := filepath.Glob(filepath.Join(cs.dataDir, "*.json"))
	if err != nil {
		return
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			continue
		}

		var session models.ChatSession
		if err := json.Unmarshal(data, &session); err != nil {
			continue
		}

		cs.sessions[session.ID] = &session
	}
}

// saveSession persists a session to disk
func (cs *ChatService) saveSession(session *models.ChatSession) error {
	data, err := json.MarshalIndent(session, "", "  ")
	if err != nil {
		return err
	}

	filename := filepath.Join(cs.dataDir, session.ID+".json")
	return os.WriteFile(filename, data, 0644)
}

// CreateSession creates a new chat session
func (cs *ChatService) CreateSession(title string) *models.ChatSession {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	now := nowMs()
	if title == "" {
		title = "New Chat"
	}

	session := &models.ChatSession{
		ID:        generateID(),
		Title:     title,
		Messages:  []models.ChatMessage{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	cs.sessions[session.ID] = session
	_ = cs.saveSession(session)

	return session
}

// GetSession retrieves a session by ID
func (cs *ChatService) GetSession(id string) *models.ChatSession {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	if session, ok := cs.sessions[id]; ok {
		return session
	}
	return nil
}

// GetAllSessions returns all sessions sorted by updated time (newest first)
func (cs *ChatService) GetAllSessions() []models.ChatSessionSummary {
	cs.mu.RLock()
	defer cs.mu.RUnlock()

	summaries := make([]models.ChatSessionSummary, 0, len(cs.sessions))
	for _, session := range cs.sessions {
		summaries = append(summaries, models.ChatSessionSummary{
			ID:           session.ID,
			Title:        session.Title,
			CreatedAt:    session.CreatedAt,
			UpdatedAt:    session.UpdatedAt,
			MessageCount: len(session.Messages),
		})
	}

	// Sort by UpdatedAt descending (newest first)
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].UpdatedAt > summaries[j].UpdatedAt
	})

	return summaries
}

// AddMessage adds a message to a session
func (cs *ChatService) AddMessage(sessionID string, role, content, riskLevel string) *models.ChatMessage {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	session, ok := cs.sessions[sessionID]
	if !ok {
		return nil
	}

	msg := models.ChatMessage{
		ID:        generateID(),
		Role:      role,
		Content:   content,
		RiskLevel: riskLevel,
		Timestamp: nowMs(),
	}

	session.Messages = append(session.Messages, msg)
	session.UpdatedAt = nowMs()

	// Auto-update title from first user message if still default
	if session.Title == "New Chat" && role == "user" && len(content) > 0 {
		title := content
		if len(title) > 50 {
			title = title[:47] + "..."
		}
		session.Title = title
	}

	_ = cs.saveSession(session)

	return &msg
}

// UpdateSessionTitle updates a session's title
func (cs *ChatService) UpdateSessionTitle(sessionID, title string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	session, ok := cs.sessions[sessionID]
	if !ok {
		return nil
	}

	session.Title = title
	session.UpdatedAt = nowMs()
	return cs.saveSession(session)
}

// DeleteSession removes a session
func (cs *ChatService) DeleteSession(sessionID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	delete(cs.sessions, sessionID)

	filename := filepath.Join(cs.dataDir, sessionID+".json")
	return os.Remove(filename)
}

// ClearSessionMessages clears all messages in a session
func (cs *ChatService) ClearSessionMessages(sessionID string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	session, ok := cs.sessions[sessionID]
	if !ok {
		return nil
	}

	session.Messages = []models.ChatMessage{}
	session.UpdatedAt = nowMs()
	return cs.saveSession(session)
}
