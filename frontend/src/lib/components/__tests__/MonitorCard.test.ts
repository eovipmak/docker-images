import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import MonitorCard from '../MonitorCard.svelte';
import type { Monitor } from '../../types';

describe('MonitorCard', () => {
  it('renders monitor content and link', () => {
    const monitor: Monitor = {
      id: 'mon-1',
      user_id: 1,
      name: 'Example Monitor',
      url: 'https://example.com',
      type: 'http',
      enabled: true,
      check_interval: 60,
      timeout: 30,
      check_ssl: true,
      ssl_alert_days: 30,
      created_at: new Date().toISOString(),
      updated_at: new Date().toISOString(),
      last_checked_at: new Date().toISOString(),
      last_check: {
          monitor_id: 'mon-1',
          checked_at: new Date().toISOString(),
          response_time_ms: 123,
          success: true
      }
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
