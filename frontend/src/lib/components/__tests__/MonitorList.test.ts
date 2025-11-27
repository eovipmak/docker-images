import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import MonitorList from '../MonitorList.svelte';

describe('MonitorList', () => {
  it('renders grid snapshot correctly', () => {
    const monitors = [
      { id: '1', name: 'A', url: 'https://a.example', enabled: true, last_checked_at: new Date().toISOString(), response_time_ms: 100 },
      { id: '2', name: 'B', url: 'https://b.example', enabled: true }
    ];

    const { container } = render(MonitorList, { props: { monitors, useTable: false } });
    expect(container).toMatchSnapshot();
  });
});
