import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import MonitorCard from '../MonitorCard.svelte';

describe('MonitorCard', () => {
  it('matches snapshot', () => {
    const monitor = {
      id: 'mon-1',
      name: 'Example Monitor',
      url: 'https://example.com',
      enabled: true,
      last_checked_at: new Date().toISOString(),
      response_time_ms: 123
    };

    const { container } = render(MonitorCard, { props: { monitor } });
    expect(container).toMatchSnapshot();
  });
});
