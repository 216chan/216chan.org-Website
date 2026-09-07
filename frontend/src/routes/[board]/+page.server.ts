import type { PageServerLoad } from './$types';

const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

export const load: PageServerLoad = async ({ params, url, fetch }) => {
	const board = params.board;
	const page = url.searchParams.get('page') ?? '1';
	try {
		const res = await fetch(`${API_BASE}/api/${board}/threads?page=${page}&limit=10`);
		if (!res.ok) throw new Error(`${res.status}`);
		const data = await res.json();
		return { board, data, error: null };
	} catch (e) {
		return { board, data: null, error: String(e) };
	}
};