import type { PageServerLoad } from './$types';

const API_BASE = process.env.API_URL ?? 'http://localhost:8080';

export const load: PageServerLoad = async ({ params, url, fetch }) => {
	const board = params.board;
	const id = params.id;
	const page = url.searchParams.get('page') ?? '1';
	try {
		const res = await fetch(`${API_BASE}/api/${board}/threads/${id}?page=${page}&limit=15`);
		if (!res.ok) throw new Error(`${res.status}`);
		const data = await res.json();
		return { board, id, data, error: null };
	} catch (e) {
		return { board, id, data: null, error: String(e) };
	}
};