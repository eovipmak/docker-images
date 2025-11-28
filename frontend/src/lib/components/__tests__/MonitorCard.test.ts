import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import MonitorCard from '../MonitorCard.svelte';

describe('MonitorCard', () => {
  it('renders monitor content and link', () => {
    const monitor = {
      id: 'mon-1',
      name: 'Example Monitor',
      url: 'https://example.com',
      enabled: true,
      last_checked_at: new Date().toISOString(),
      response_time_ms: 123
    };

    const { getByText, container } = render(MonitorCard, { props: { monitor } });
    // Verify title and url are rendered
    expect(getByText(monitor.name)).toBeTruthy();
    expect(getByText(monitor.url)).toBeTruthy();
    // verify the card has a data-testid and anchor to monitor details
    const card = container.querySelector('[data-testid="monitor-card"]');
    expect(card).toBeTruthy();
    const anchor = card?.querySelector('a');
    expect(anchor).toBeTruthy();
    // href should include the monitor id
    expect(anchor?.getAttribute('href')).toContain(`/monitors/${monitor.id}`);
  });
});
