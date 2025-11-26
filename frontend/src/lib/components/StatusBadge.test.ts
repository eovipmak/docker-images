import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import StatusBadge from './StatusBadge.svelte';

describe('StatusBadge', () => {
	describe('status rendering', () => {
		it('should render "Open" for open status', () => {
			render(StatusBadge, { props: { status: 'open' } });
			expect(screen.getByText('Open')).toBeTruthy();
		});

		it('should render "Resolved" for resolved status', () => {
			render(StatusBadge, { props: { status: 'resolved' } });
			expect(screen.getByText('Resolved')).toBeTruthy();
		});
	});

	describe('styling', () => {
		it('should apply red colors for open status', () => {
			const { container } = render(StatusBadge, { props: { status: 'open' } });
			const badge = container.querySelector('span');
			
			expect(badge?.classList.contains('bg-red-100')).toBe(true);
			expect(badge?.classList.contains('text-red-700')).toBe(true);
		});

		it('should apply green colors for resolved status', () => {
			const { container } = render(StatusBadge, { props: { status: 'resolved' } });
			const badge = container.querySelector('span');
			
			expect(badge?.classList.contains('bg-green-100')).toBe(true);
			expect(badge?.classList.contains('text-green-700')).toBe(true);
		});
	});

	describe('size variants', () => {
		it('should apply small size classes when size="sm"', () => {
			const { container } = render(StatusBadge, { props: { status: 'open', size: 'sm' } });
			const badge = container.querySelector('span');
			
			expect(badge?.classList.contains('px-2')).toBe(true);
			expect(badge?.classList.contains('py-0.5')).toBe(true);
			expect(badge?.classList.contains('text-xs')).toBe(true);
		});

		it('should apply medium size classes when size="md"', () => {
			const { container } = render(StatusBadge, { props: { status: 'open', size: 'md' } });
			const badge = container.querySelector('span');
			
			expect(badge?.classList.contains('px-2.5')).toBe(true);
			expect(badge?.classList.contains('py-0.5')).toBe(true);
			expect(badge?.classList.contains('text-sm')).toBe(true);
		});

		it('should default to medium size when size prop is not provided', () => {
			const { container } = render(StatusBadge, { props: { status: 'open' } });
			const badge = container.querySelector('span');
			
			expect(badge?.classList.contains('px-2.5')).toBe(true);
			expect(badge?.classList.contains('text-sm')).toBe(true);
		});
	});

	describe('base styling', () => {
		it('should always have base badge classes', () => {
			const { container } = render(StatusBadge, { props: { status: 'open' } });
			const badge = container.querySelector('span');
			
			expect(badge?.classList.contains('inline-flex')).toBe(true);
			expect(badge?.classList.contains('items-center')).toBe(true);
			expect(badge?.classList.contains('rounded-full')).toBe(true);
			expect(badge?.classList.contains('font-medium')).toBe(true);
		});
	});
});
