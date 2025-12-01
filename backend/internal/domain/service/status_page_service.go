package service

import (
	"fmt"
	"regexp"

	"github.com/eovipmak/v-insight/backend/internal/database"
	"github.com/eovipmak/v-insight/backend/internal/domain/entities"
	"github.com/eovipmak/v-insight/backend/internal/domain/repository"
)

// MonitorWithStatus represents a monitor with its current status for public display
type MonitorWithStatus struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
	Status  string `json:"status"` // "up", "down", "unknown"
}

// StatusPageService provides business logic for status page operations
type StatusPageService struct {
	db         *database.DB
	repo       repository.StatusPageRepository
	monitorRepo repository.MonitorRepository
}
func NewStatusPageService(db *database.DB, repo repository.StatusPageRepository, monitorRepo repository.MonitorRepository) *StatusPageService {
	return &StatusPageService{
		db:         db,
		repo:       repo,
		monitorRepo: monitorRepo,
	}
}

// CreateStatusPage creates a new status page
func (s *StatusPageService) CreateStatusPage(tenantID int, slug, name string, publicEnabled bool) (*entities.StatusPage, error) {
	// Validate slug format
	if !isValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format")
	}

	// Check if slug is unique
	existing, err := s.repo.GetBySlug(slug)
	if err == nil && existing != nil {
		return nil, fmt.Errorf("slug already exists")
	}

	statusPage := &entities.StatusPage{
		TenantID:      tenantID,
		Slug:          slug,
		Name:          name,
		PublicEnabled: publicEnabled,
	}

	err = s.repo.Create(statusPage)
	if err != nil {
		return nil, fmt.Errorf("failed to create status page: %w", err)
	}

	return statusPage, nil
}

// GetStatusPageByID retrieves a status page by ID
func (s *StatusPageService) GetStatusPageByID(id string, tenantID int) (*entities.StatusPage, error) {
	statusPage, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get status page: %w", err)
	}

	if statusPage.TenantID != tenantID {
		return nil, fmt.Errorf("status page not found")
	}

	return statusPage, nil
}

// GetStatusPagesByTenant retrieves all status pages for a tenant
func (s *StatusPageService) GetStatusPagesByTenant(tenantID int) ([]*entities.StatusPage, error) {
	statusPages, err := s.repo.GetByTenantID(tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get status pages: %w", err)
	}

	return statusPages, nil
}

// UpdateStatusPage updates an existing status page
func (s *StatusPageService) UpdateStatusPage(id string, tenantID int, slug, name string, publicEnabled bool) (*entities.StatusPage, error) {
	statusPage, err := s.GetStatusPageByID(id, tenantID)
	if err != nil {
		return nil, err
	}

	// Validate slug format
	if !isValidSlug(slug) {
		return nil, fmt.Errorf("invalid slug format")
	}

	// Check if slug is unique (excluding current)
	existing, err := s.repo.GetBySlug(slug)
	if err == nil && existing != nil && existing.ID != id {
		return nil, fmt.Errorf("slug already exists")
	}

	statusPage.Slug = slug
	statusPage.Name = name
	statusPage.PublicEnabled = publicEnabled

	err = s.repo.Update(statusPage)
	if err != nil {
		return nil, fmt.Errorf("failed to update status page: %w", err)
	}

	return statusPage, nil
}

// DeleteStatusPage deletes a status page
func (s *StatusPageService) DeleteStatusPage(id string, tenantID int) error {
	_, err := s.GetStatusPageByID(id, tenantID)
	if err != nil {
		return err
	}

	return s.repo.Delete(id)
}

// AddMonitorToStatusPage adds a monitor to a status page
func (s *StatusPageService) AddMonitorToStatusPage(statusPageID, monitorID string, tenantID int) error {
	// Verify status page belongs to tenant
	_, err := s.GetStatusPageByID(statusPageID, tenantID)
	if err != nil {
		return err
	}

	// Verify monitor belongs to tenant
	monitor, err := s.monitorRepo.GetByID(monitorID)
	if err != nil {
		return fmt.Errorf("monitor not found: %w", err)
	}
	if monitor.TenantID != tenantID {
		return fmt.Errorf("monitor not found")
	}

	return s.repo.AddMonitor(statusPageID, monitorID)
}

// RemoveMonitorFromStatusPage removes a monitor from a status page
func (s *StatusPageService) RemoveMonitorFromStatusPage(statusPageID, monitorID string, tenantID int) error {
	// Verify status page belongs to tenant
	_, err := s.GetStatusPageByID(statusPageID, tenantID)
	if err != nil {
		return err
	}

	return s.repo.RemoveMonitor(statusPageID, monitorID)
}

// GetStatusPageMonitors retrieves monitors for a status page
func (s *StatusPageService) GetStatusPageMonitors(statusPageID string, tenantID int) ([]*entities.Monitor, error) {
	// Verify status page belongs to tenant
	_, err := s.GetStatusPageByID(statusPageID, tenantID)
	if err != nil {
		return nil, err
	}

	return s.repo.GetMonitors(statusPageID)
}

// GetPublicStatusPage retrieves a public status page by slug
func (s *StatusPageService) GetPublicStatusPage(slug string) (*entities.StatusPage, []*MonitorWithStatus, error) {
	statusPage, err := s.repo.GetBySlug(slug)
	if err != nil {
		return nil, nil, fmt.Errorf("status page not found: %w", err)
	}

	if !statusPage.PublicEnabled {
		return nil, nil, fmt.Errorf("status page not public")
	}

	monitors, err := s.repo.GetMonitors(statusPage.ID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get monitors: %w", err)
	}

	// For each monitor, determine current status from latest check
	monitorsWithStatus := make([]*MonitorWithStatus, len(monitors))
	for i, monitor := range monitors {
		status := "unknown"
		if monitor.Enabled {
			// Get the latest check
			checks, err := s.monitorRepo.GetChecksByMonitorID(monitor.ID, 1)
			if err == nil && len(checks) > 0 {
				if checks[0].Success {
					status = "up"
				} else {
					status = "down"
				}
			}
		}

		monitorsWithStatus[i] = &MonitorWithStatus{
			ID:      monitor.ID,
			Name:    monitor.Name,
			URL:     monitor.URL,
			Type:    monitor.Type,
			Enabled: monitor.Enabled,
			Status:  status,
		}
	}

	return statusPage, monitorsWithStatus, nil
}

// isValidSlug validates the slug format (alphanumeric, hyphens, underscores)
func isValidSlug(slug string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_-]+$`, slug)
	return matched && len(slug) > 0 && len(slug) <= 255
}