import type { PageLoad } from './$types';

interface MonitorWithStatus {
	id: string;
	name: string;
	url: string;
	type: string;
	enabled: boolean;
	status: string; // "up", "down", "unknown"
}

interface StatusPage {
	id: string;
	user_id: number;
	slug: string;
	name: string;
	public_enabled: boolean;
	created_at: string;
	updated_at: string;
}

interface StatusPageData {
	status_page: StatusPage;
	monitors: MonitorWithStatus[];
}

export const load: PageLoad = async ({ params, fetch }) => {
	try {
		const response = await fetch(`/api/public/status/${params.slug}`);
		if (!response.ok) {
			throw new Error('Status page not found');
		}
		const data: StatusPageData = await response.json();
		return {
			statusPage: data.status_page,
			monitors: data.monitors
		};
	} catch (error) {
		throw new Error('Status page not found');
	}
};