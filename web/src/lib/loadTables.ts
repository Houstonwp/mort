import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';

import type { ConvertedTable, TableIndexEntry, TableSummary } from './types';

const jsonDirUrl = new URL('../../../json/', import.meta.url);
const jsonDirPath = fileURLToPath(jsonDirUrl);

let cachedIndex: TableIndexEntry[] | null = null;

export async function loadTableIndex(): Promise<TableIndexEntry[]> {
  if (cachedIndex) {
    return cachedIndex;
  }

  const entries = await fs.readdir(jsonDirPath, { withFileTypes: true });
  const files = entries.filter((entry) => entry.isFile() && entry.name.endsWith('.json'));

  const tables = await Promise.all(files.map(async (entry) => {
    const absolutePath = path.join(jsonDirPath, entry.name);
    const raw = await fs.readFile(absolutePath, 'utf-8');
    const detail = JSON.parse(raw) as ConvertedTable;
    return toIndexEntry(detail, absolutePath, entry.name);
  }));

  tables.sort((a, b) => compareIdentities(a.tableIdentity, b.tableIdentity, a.name, b.name));
  cachedIndex = tables;
  return tables;
}

export async function loadTableDetail(filePath: string): Promise<ConvertedTable> {
  const raw = await fs.readFile(filePath, 'utf-8');
  return JSON.parse(raw) as ConvertedTable;
}

function toIndexEntry(detail: ConvertedTable, filePath: string, fileName: string): TableIndexEntry {
  const classification = detail.classification;
  const tableIdentity = classification?.tableIdentity ?? detail.identifier ?? fileName.replace(/\.json$/i, '');
  const name = classification?.tableName ?? detail.identifier ?? tableIdentity;
  const provider = classification?.providerName ?? 'Unknown provider';
  const summary = classification?.tableDescription ?? classification?.comments ?? '';
  const keywords = classification?.keywords ?? [];
  const identifier = detail.identifier ?? tableIdentity;
  const version = detail.version ?? '';
  const detailPath = `/detail/${encodeURIComponent(identifier)}.json`;

  const summaryRecord: TableSummary = {
    identifier,
    tableIdentity,
    name,
    provider,
    summary,
    keywords,
    version,
    detailPath,
  };

  return {
    ...summaryRecord,
    filePath,
  };
}

function compareIdentities(aId: string, bId: string, aName: string, bName: string): number {
  const [aNum, aIsNumber] = parseIdentity(aId);
  const [bNum, bIsNumber] = parseIdentity(bId);

  if (aIsNumber && bIsNumber && aNum !== bNum) {
    return aNum - bNum;
  }
  if (aIsNumber && !bIsNumber) {
    return -1;
  }
  if (!aIsNumber && bIsNumber) {
    return 1;
  }
  if (aId !== bId) {
    return aId.localeCompare(bId);
  }
  return aName.localeCompare(bName);
}

function parseIdentity(id: string): [number, boolean] {
  const numeric = Number(id);
  if (!Number.isNaN(numeric)) {
    return [numeric, true];
  }
  return [0, false];
}
