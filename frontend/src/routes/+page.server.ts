import type { PageServerLoad } from './$types';

const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

export const load: PageServerLoad = async ({ fetch }) => {
	try {
		const res = await fetch(`${API_BASE}/api/stats`);
		if (!res.ok) throw new Error(`stats ${res.status}`);
		const stats = await res.json();
		return { stats };
	} catch {
		return {
			stats: {
				total_posts: 0,
				current_users: 0,
				active_content: '0 B'
			}
		};
	}
};
