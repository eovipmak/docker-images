import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import SSLCheckForm from '../components/SSLCheckForm';
import { LanguageProvider } from '../hooks/useLanguage';

// Mock the API module
vi.mock('../services/api', () => ({
  addDomain: vi.fn(),
  parseTarget: vi.fn((target: string) => ({ host: target, port: 443 })),
}));

const renderWithLanguage = (component: React.ReactElement) => {
  return render(<LanguageProvider>{component}</LanguageProvider>);
};

describe('SSLCheckForm', () => {
  it('renders the form with input and button', () => {
    const mockOnResult = vi.fn();
    renderWithLanguage(<SSLCheckForm onResult={mockOnResult} />);

    // Check for form elements (English labels)
    expect(screen.getByRole('textbox', { name: /domain or ip address/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /check certificate/i })).toBeInTheDocument();
  });

  it('displays validation error when submitting empty form', async () => {
    const mockOnResult = vi.fn();
    renderWithLanguage(<SSLCheckForm onResult={mockOnResult} />);

    const submitButton = screen.getByRole('button', { name: /check certificate/i });
    
    // Click submit without entering a domain
    fireEvent.click(submitButton);

    // Wait for validation error (English text)
    await waitFor(() => {
      expect(screen.getByText(/please provide a domain name or ip address/i)).toBeInTheDocument();
    });

    // Should not call onResult
    expect(mockOnResult).not.toHaveBeenCalled();
  });

  it('allows user to enter a domain name', async () => {
    const user = userEvent.setup();
    const mockOnResult = vi.fn();
    renderWithLanguage(<SSLCheckForm onResult={mockOnResult} />);

    const input = screen.getByRole('textbox', { name: /domain or ip address/i });
    
    // Type a domain name
    await user.type(input, 'example.com');

    // Check that the input value is updated
    expect(input).toHaveValue('example.com');
  });

  it('has accessible form elements', () => {
    const mockOnResult = vi.fn();
    renderWithLanguage(<SSLCheckForm onResult={mockOnResult} />);

    const input = screen.getByRole('textbox', { name: /domain or ip address/i });
    const button = screen.getByRole('button', { name: /check certificate/i });

    // Check for aria attributes
    expect(input).toHaveAttribute('aria-label');
    expect(input).toHaveAttribute('aria-required', 'true');
    expect(button).toHaveAttribute('aria-label');
  });
});
