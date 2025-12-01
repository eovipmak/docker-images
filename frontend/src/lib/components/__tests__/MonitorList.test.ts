import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import MonitorList from '../MonitorList.svelte';
import type { Monitor } from '../../types';

describe('MonitorList', () => {
  it('renders grid snapshot correctly', () => {
    const monitors: Monitor[] = [
      {
        id: '1',
        tenant_id: 1,
        name: 'A',
        url: 'https://a.example',
        type: 'http',
        enabled: true,
        check_interval: 60,
        timeout: 30,
        check_ssl: true,
        ssl_alert_days: 30,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
        last_checked_at: new Date().toISOString(),
        last_check: { monitor_id: '1', checked_at: new Date().toISOString(), response_time_ms: 100, success: true }
      },
      {
        id: '2',
        tenant_id: 1,
        name: 'B',
        url: 'https://b.example',
        type: 'http',
        enabled: true,
        check_interval: 60,
        timeout: 30,
        check_ssl: true,
        ssl_alert_days: 30,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString()
      }
    ];

    const { container } = render(MonitorList, { props: { monitors, useTable: false } });
    expect(container).toMatchSnapshot();
  });
});
