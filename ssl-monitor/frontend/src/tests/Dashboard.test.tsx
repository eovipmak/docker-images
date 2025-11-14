import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import Dashboard from '../pages/Dashboard';
import { LanguageProvider } from '../hooks/useLanguage';
import * as api from '../services/api';

// Mock WebSocket
class MockWebSocket {
  onopen: ((event: Event) => void) | null = null;
  onmessage: ((event: MessageEvent) => void) | null = null;
  onerror: ((event: Event) => void) | null = null;
  onclose: ((event: CloseEvent) => void) | null = null;
  
  send = vi.fn();
  close = vi.fn();
  
  constructor(public url: string) {
    setTimeout(() => {
      if (this.onopen) {
        this.onopen(new Event('open'));
      }
    }, 0);
  }
}

global.WebSocket = MockWebSocket as unknown as typeof WebSocket;

// Mock fetch for /api/domains/status
global.fetch = vi.fn();

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
global.localStorage = localStorageMock as Storage;

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
    localStorageMock.getItem.mockReturnValue('mock-token');
    (global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => ({ domains: [] }),
    });
  });

  afterEach(() => {
    vi.clearAllTimers();
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

  it('displays monitored domains with color-coded status', async () => {
    const mockStats = {
      stats: {
        total_checks: 10,
        successful_checks: 8,
        error_checks: 2,
        unique_domains: 3,
      },
    };

    const mockHistory = {
      history: [],
    };

    const mockDomains = {
      domains: [
        {
          domain: 'example.com',
          ip: '93.184.216.34',
          port: 443,
          status: 'success',
          ssl_status: 'success',
          last_checked: '2024-01-15T10:30:00Z',
          ssl_info: {
            daysUntilExpiration: 45,
            issuer: {
              organizationName: 'Let\'s Encrypt',
            },
          },
        },
        {
          domain: 'warning.com',
          ip: '1.2.3.4',
          port: 443,
          status: 'success',
          ssl_status: 'success',
          last_checked: '2024-01-15T10:30:00Z',
          ssl_info: {
            daysUntilExpiration: 15,
          },
        },
        {
          domain: 'critical.com',
          ip: '5.6.7.8',
          port: 443,
          status: 'success',
          ssl_status: 'success',
          last_checked: '2024-01-15T10:30:00Z',
          ssl_info: {
            daysUntilExpiration: 3,
          },
        },
      ],
    };

    vi.mocked(api.getStats).mockResolvedValue(mockStats);
    vi.mocked(api.getHistory).mockResolvedValue(mockHistory);
    (global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => mockDomains,
    });

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText('example.com')).toBeInTheDocument();
      expect(screen.getByText('warning.com')).toBeInTheDocument();
      expect(screen.getByText('critical.com')).toBeInTheDocument();
      expect(screen.getByText('45 days')).toBeInTheDocument();
      expect(screen.getByText('15 days')).toBeInTheDocument();
      expect(screen.getByText('3 days')).toBeInTheDocument();
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
    });
  });

  it('displays message when no domains are being monitored', async () => {
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
    (global.fetch as unknown as ReturnType<typeof vi.fn>).mockResolvedValue({
      ok: true,
      json: async () => ({ domains: [] }),
    });

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      expect(screen.getByText(/No domains are being monitored/i)).toBeInTheDocument();
    });
  });

  it('handles non-array history gracefully', async () => {
    const mockStats = {
      stats: {
        total_checks: 5,
        successful_checks: 5,
        error_checks: 0,
        unique_domains: 2,
      },
    };

    // Mock history with non-array value
    const mockHistory = {
      history: null as unknown,
    };

    vi.mocked(api.getStats).mockResolvedValue(mockStats);
    vi.mocked(api.getHistory).mockResolvedValue(mockHistory as typeof mockHistory & { history: unknown[] });

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      // Should display stats without crashing - check for total_checks
      expect(screen.getAllByText('5').length).toBeGreaterThan(0);
      // Should show "No checks found" instead of crashing
      expect(screen.getByText(/No checks found/i)).toBeInTheDocument();
    });
  });

  it('handles missing history property gracefully', async () => {
    const mockStats = {
      stats: {
        total_checks: 7,
        successful_checks: 6,
        error_checks: 1,
        unique_domains: 3,
      },
    };

    // Mock history without history property
    const mockHistory = {} as { history: unknown[] };

    vi.mocked(api.getStats).mockResolvedValue(mockStats);
    vi.mocked(api.getHistory).mockResolvedValue(mockHistory);

    renderWithLanguage(<Dashboard />);

    await waitFor(() => {
      // Should display stats without crashing - check for unique value
      expect(screen.getByText('7')).toBeInTheDocument();
      // Should show "No checks found" instead of crashing
      expect(screen.getByText(/No checks found/i)).toBeInTheDocument();
    });
  });
});
