import type { APIRoute, GetStaticPaths } from 'astro';
import { loadTableDetail, loadTableIndex } from '../../lib/loadTables';

export const prerender = true;

export const getStaticPaths: GetStaticPaths = async () => {
  const index = await loadTableIndex();
  return index.map((entry) => ({
    params: { identifier: entry.identifier },
    props: { filePath: entry.filePath },
  }));
};

export const GET: APIRoute = async ({ props }) => {
  const detail = await loadTableDetail(props.filePath as string);
  return new Response(JSON.stringify(detail), {
    status: 200,
    headers: { 'Content-Type': 'application/json' },
  });
};
