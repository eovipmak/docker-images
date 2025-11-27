import { describe, it, expect } from 'vitest';
import { render } from '@testing-library/svelte';
import AlertCard from '../AlertCard.svelte';

describe('AlertCard', () => {
  it('matches snapshot', () => {
    const rule = {
      id: 'rule-1',
      name: 'Down rule',
      trigger_type: 'down',
      monitor_id: null,
      threshold_value: 3,
      enabled: true
    };

    const { container } = render(AlertCard, { props: { rule } });
    expect(container).toMatchSnapshot();
  });
});
