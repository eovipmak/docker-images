import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import Dashboard from '../pages/Dashboard';
import { LanguageProvider } from '../hooks/useLanguage';
import * as api from '../services/api';

// Mock the API module
vi.mock('../services/api', () => ({
  getStats: vi.fn(),
  getHistory: vi.fn(),
}));

const renderWithLanguage = (component: React.ReactElement) => {
  return render(<LanguageProvider>{component}</LanguageProvider>);
};

describe('Dashboard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('displays loading state initially', () => {
    // Mock API calls that never resolve
    vi.mocked(api.getStats).mockImplementation(() => new Promise(() => {}));
    vi.mocked(api.getHistory).mockImplementation(() => new Promise(() => {}));

    renderWithLanguage(<Dashboard />);
    
    expect(screen.getByRole('progressbar')).toBeInTheDocument();
  });

  it('displays error message when API fails', async () => {
    const errorMessage = 'Failed to fetch data';
    vi.mocked(api.getStats).mockRejectedValue(new Error(errorMessage));
    vi.mocked(api.getHistory).mockRejectedValue(new Error(errorMessage));

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument();
    });
  });

  it('displays stats from API', async () => {
    const mockStats = {
      stats: {
        total_checks: 100,
        successful_checks: 80,
        error_checks: 20,
        unique_domains: 25,
      },
    };

    const mockHistory = {
      history: [],
    };

    vi.mocked(api.getStats).mockResolvedValue(mockStats);
    vi.mocked(api.getHistory).mockResolvedValue(mockHistory);

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('100')).toBeInTheDocument();
      expect(screen.getByText('80')).toBeInTheDocument();
      expect(screen.getByText('20')).toBeInTheDocument();
      expect(screen.getByText('25')).toBeInTheDocument();
    });
  });

  it('displays recent checks from API', async () => {
    const mockStats = {
      stats: {
        total_checks: 10,
        successful_checks: 8,
        error_checks: 2,
        unique_domains: 5,
      },
    };

    const mockHistory = {
      history: [
        {
          id: 1,
          domain: 'example.com',
          status: 'success',
          checked_at: '2024-01-15T10:30:00Z',
          ssl_status: 'valid',
          data: {
            ssl: {
              daysUntilExpiration: 45,
            },
          },
        },
        {
          id: 2,
          domain: 'test.com',
          status: 'error',
          checked_at: '2024-01-15T09:15:00Z',
          ssl_status: 'invalid',
        },
      ],
    };

    vi.mocked(api.getStats).mockResolvedValue(mockStats);
    vi.mocked(api.getHistory).mockResolvedValue(mockHistory);

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('example.com')).toBeInTheDocument();
      expect(screen.getByText('test.com')).toBeInTheDocument();
      expect(screen.getByText('45 days')).toBeInTheDocument();
    });
  });

  it('displays message when no checks are available', async () => {
    const mockStats = {
      stats: {
        total_checks: 0,
        successful_checks: 0,
        error_checks: 0,
        unique_domains: 0,
      },
    };

    const mockHistory = {
      history: [],
    };

    vi.mocked(api.getStats).mockResolvedValue(mockStats);
    vi.mocked(api.getHistory).mockResolvedValue(mockHistory);

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText(/No checks found/i)).toBeInTheDocument();
    });
  });
});
