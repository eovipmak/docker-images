import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import AlertsDisplay from '../components/AlertsDisplay';
import { LanguageProvider } from '../hooks/useLanguage';
import * as api from '../services/api';

// Mock the API module
vi.mock('../services/api', () => ({
  getAlerts: vi.fn(),
  markAlertRead: vi.fn(),
  markAlertResolved: vi.fn(),
  deleteAlert: vi.fn(),
}));

const renderWithLanguage = (component: React.ReactElement) => {
  return render(<LanguageProvider>{component}</LanguageProvider>);
};

describe('AlertsDisplay', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('handles non-array response gracefully', async () => {
    // Mock API returning a non-array value (simulating an error case)
    vi.mocked(api.getAlerts).mockResolvedValue(null as any);

    renderWithLanguage(<AlertsDisplay />);

    await waitFor(() => {
      // Should show "No alerts to display" instead of crashing
      expect(screen.getByText(/No alerts to display/i)).toBeInTheDocument();
    });
  });

  it('handles undefined response gracefully', async () => {
    // Mock API returning undefined
    vi.mocked(api.getAlerts).mockResolvedValue(undefined as any);

    renderWithLanguage(<AlertsDisplay />);

    await waitFor(() => {
      // Should show "No alerts to display" instead of crashing
      expect(screen.getByText(/No alerts to display/i)).toBeInTheDocument();
    });
  });

  it('handles object response gracefully', async () => {
    // Mock API returning an object instead of an array
    vi.mocked(api.getAlerts).mockResolvedValue({ alerts: [] } as any);

    renderWithLanguage(<AlertsDisplay />);

    await waitFor(() => {
      // Should show "No alerts to display" instead of crashing
      expect(screen.getByText(/No alerts to display/i)).toBeInTheDocument();
    });
  });

  it('displays alerts when API returns valid array', async () => {
    const mockAlerts = [
      {
        id: 1,
        user_id: 1,
        domain: 'example.com',
        alert_type: 'expiring_soon' as const,
        severity: 'medium' as const,
        message: 'Certificate expiring in 15 days',
        is_read: false,
        is_resolved: false,
        created_at: '2024-01-15T10:30:00Z',
      },
    ];

    vi.mocked(api.getAlerts).mockResolvedValue(mockAlerts);

    renderWithLanguage(<AlertsDisplay />);

    await waitFor(() => {
      expect(screen.getByText('example.com')).toBeInTheDocument();
      expect(screen.getByText(/Certificate expiring in 15 days/i)).toBeInTheDocument();
    });
  });

  it('displays empty state when API returns empty array', async () => {
    vi.mocked(api.getAlerts).mockResolvedValue([]);

    renderWithLanguage(<AlertsDisplay />);

    await waitFor(() => {
      expect(screen.getByText(/No alerts to display/i)).toBeInTheDocument();
    });
  });
});
