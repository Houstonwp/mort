import { useEffect, useMemo, useRef, useState } from 'preact/hooks';
import Fuse from 'fuse.js';
import JSZip from 'jszip';

import type { ConvertedTable, DetailTab, TablePayload, TableSummary } from '../lib/types';

const LOAD_BATCH = 40;
const TAB_ORDER: DetailTab[] = ['classification', 'metadata', 'rates'];
type RateView = 'list' | 'matrix';

interface MortalityAppProps {
  tables: TableSummary[];
}

export default function MortalityApp({ tables }: MortalityAppProps) {
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<TableSummary | null>(null);
  const [detail, setDetail] = useState<ConvertedTable | null>(null);
  const [detailState, setDetailState] = useState<'idle' | 'loading' | 'error'>('idle');
  const [activeTab, setActiveTab] = useState<DetailTab>('classification');
  const [tableIndex, setTableIndex] = useState(0);
  const [rateView, setRateView] = useState<RateView>('list');
  const [visibleCount, setVisibleCount] = useState(LOAD_BATCH);
  const listRef = useRef<HTMLDivElement | null>(null);
  const [csvDownloadingId, setCsvDownloadingId] = useState<string | null>(null);
  const [selectedKeys, setSelectedKeys] = useState<Set<string>>(() => new Set());
  const [bulkLoading, setBulkLoading] = useState<'json' | 'csv' | null>(null);
  const [lastSelectedIndex, setLastSelectedIndex] = useState<number | null>(null);
  const selectAllRef = useRef<HTMLInputElement | null>(null);

  const fuse = useMemo(() => {
    return new Fuse(tables, {
      keys: ['name', 'identifier', 'tableIdentity', 'provider', 'summary', 'keywords'],
      threshold: 0.32,
      ignoreLocation: true,
    });
  }, [tables]);

  const filtered = useMemo(() => {
    const search = query.trim();
    if (!search) {
      return tables;
    }
    return fuse.search(search).map((result) => result.item);
  }, [fuse, query, tables]);

  const visibleTables = filtered.slice(0, visibleCount);
  const hasMore = visibleCount < filtered.length;
  const selectedSummaries = useMemo(
    () => tables.filter((table) => selectedKeys.has(table.detailPath)),
    [tables, selectedKeys],
  );
  const selectedCount = selectedKeys.size;
  const visibleSelectedCount = useMemo(
    () => visibleTables.filter((table) => selectedKeys.has(table.detailPath)).length,
    [visibleTables, selectedKeys],
  );
  const filteredSelectedCount = useMemo(
    () => filtered.filter((table) => selectedKeys.has(table.detailPath)).length,
    [filtered, selectedKeys],
  );
  const triggerJsonDownload = (detailPath: string, identifier?: string) => {
    const link = document.createElement('a');
    link.href = detailPath;
    if (identifier) {
      link.download = `${identifier}.json`;
    } else {
      link.setAttribute('download', '');
    }
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  const triggerCsvDownload = (tableDetail: ConvertedTable, indexes: number[] | 'all') => {
    const indices =
      indexes === 'all'
        ? tableDetail.tables?.map((_, idx) => idx) ?? []
        : indexes.filter((idx) => tableDetail.tables && idx >= 0 && idx < tableDetail.tables.length);
    indices.forEach((idx) => {
      const csv = buildCsv(tableDetail, idx);
      if (!csv) {
        return;
      }
      const blob = new Blob([csv], { type: 'text/csv;charset=utf-8;' });
      const url = URL.createObjectURL(blob);
      const filename = `${tableDetail.identifier || 'table'}_${(tableDetail.tables?.[idx]?.index ?? idx) + 1}.csv`;
      const link = document.createElement('a');
      link.href = url;
      link.download = filename;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    });
  };

  const handleDetailCsvDownload = () => {
    if (!detail) {
      return;
    }
    const hasMultiple = (detail.tables?.length ?? 0) > 1;
    if (hasMultiple) {
      triggerCsvDownload(detail, 'all');
    } else {
      triggerCsvDownload(detail, [tableIndex]);
    }
  };

  const handleSummaryCsvDownload = async (summary: TableSummary) => {
    setCsvDownloadingId(summary.detailPath);
    try {
      const resp = await fetch(summary.detailPath);
      if (!resp.ok) {
        throw new Error('Failed to fetch detail');
      }
      const data = (await resp.json()) as ConvertedTable;
      const hasMultiple = (data.tables?.length ?? 0) > 1;
      if (hasMultiple) {
        triggerCsvDownload(data, 'all');
      } else {
        triggerCsvDownload(data, [0]);
      }
    } catch (err) {
      console.error('csv download failed', err);
    } finally {
      setCsvDownloadingId(null);
    }
  };

  const handleCheckboxSelection = (
    detailPath: string,
    options: { shiftKey: boolean; nextChecked: boolean },
  ) => {
    const currentIndex = filtered.findIndex((table) => table.detailPath === detailPath);
    if (currentIndex === -1) {
      return;
    }
    setSelectedKeys((prev) => {
      const next = new Set(prev);
      const applySelection = (path: string) => {
        if (options.nextChecked) {
          next.add(path);
        } else {
          next.delete(path);
        }
      };
      if (options.shiftKey && lastSelectedIndex !== null) {
        const start = Math.min(currentIndex, lastSelectedIndex);
        const end = Math.max(currentIndex, lastSelectedIndex);
        for (let idx = start; idx <= end; idx++) {
          const rangePath = filtered[idx]?.detailPath;
          if (rangePath) {
            applySelection(rangePath);
          }
        }
      } else {
        applySelection(detailPath);
      }
      return next;
    });
    setLastSelectedIndex(currentIndex);
  };

  const handleSelectAllFiltered = () => {
    setSelectedKeys((prev) => {
      const next = new Set(prev);
      const allSelected = filtered.length > 0 && filtered.every((table) => next.has(table.detailPath));
      if (allSelected) {
        filtered.forEach((table) => next.delete(table.detailPath));
      } else {
        filtered.forEach((table) => next.add(table.detailPath));
      }
      return next;
    });
  };

  const clearSelection = () => setSelectedKeys(new Set());

  const handleBulkDownload = async (kind: 'json' | 'csv') => {
    if (selectedSummaries.length === 0) {
      return;
    }
    setBulkLoading(kind);
    try {
      const zip = new JSZip();
      for (const summary of selectedSummaries) {
        const resp = await fetch(summary.detailPath);
        if (!resp.ok) {
          throw new Error(`Failed to fetch ${summary.detailPath}`);
        }
        if (kind === 'json') {
          const text = await resp.text();
          const filename = `${summary.identifier || summary.tableIdentity}.json`;
          zip.file(filename, text);
        } else {
          const data = (await resp.json()) as ConvertedTable;
          const tableCount = data.tables?.length ?? 0;
          if (tableCount === 0) {
            continue;
          }
          for (let idx = 0; idx < tableCount; idx++) {
            const csv = buildCsv(data, idx);
            if (!csv) {
              continue;
            }
            const filename = `${data.identifier || summary.tableIdentity}_table-${(
              data.tables?.[idx]?.index ?? idx
            ) + 1}.csv`;
            zip.file(filename, csv);
          }
        }
      }
      const blob = await zip.generateAsync({ type: 'blob' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = kind === 'json' ? 'tables-json.zip' : 'tables-csv.zip';
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    } catch (err) {
      console.error('bulk download failed', err);
    } finally {
      setBulkLoading(null);
    }
  };

  useEffect(() => {
    setVisibleCount(Math.min(LOAD_BATCH, filtered.length));
    if (listRef.current) {
      listRef.current.scrollTop = 0;
    }
  }, [query, filtered.length]);

  useEffect(() => {
    setVisibleCount((prev) => Math.min(prev, filtered.length));
  }, [filtered.length]);

  useEffect(() => {
    setLastSelectedIndex(null);
  }, [filtered]);

  useEffect(() => {
    const node = listRef.current;
    if (!node) {
      return;
    }
    const handleScroll = () => {
      if (node.scrollTop + node.clientHeight >= node.scrollHeight - 48) {
        setVisibleCount((prev) => Math.min(filtered.length, prev + LOAD_BATCH));
      }
    };
    node.addEventListener('scroll', handleScroll);
    return () => {
      node.removeEventListener('scroll', handleScroll);
    };
  }, [filtered.length]);

  useEffect(() => {
    if (typeof document === 'undefined') {
      return;
    }
    const body = document.body;
    if (selected) {
      body.classList.add('modal-open');
      return () => {
        body.classList.remove('modal-open');
      };
    }
    body.classList.remove('modal-open');
  }, [selected]);

  useEffect(() => {
    if (!selectAllRef.current) {
      return;
    }
    const totalFiltered = filtered.length;
    selectAllRef.current.indeterminate =
      filteredSelectedCount > 0 && filteredSelectedCount < totalFiltered;
    selectAllRef.current.checked = totalFiltered > 0 && filteredSelectedCount === totalFiltered;
  }, [filtered.length, filteredSelectedCount]);

  useEffect(() => {
    if (!selected) {
      setDetail(null);
      setDetailState('idle');
      return;
    }
    let canceled = false;
    setDetailState('loading');
    setActiveTab('classification');
    setTableIndex(0);

    fetch(selected.detailPath)
      .then((resp) => {
        if (!resp.ok) {
          throw new Error('Failed to load detail');
        }
        return resp.json() as Promise<ConvertedTable>;
      })
      .then((data) => {
        if (canceled) return;
        setDetail(data);
        setDetailState('idle');
      })
      .catch(() => {
        if (canceled) return;
        setDetail(null);
        setDetailState('error');
      });

    return () => {
      canceled = true;
    };
  }, [selected]);

  useEffect(() => {
    if (!detail) {
      if (rateView !== 'list') {
        setRateView('list');
      }
      return;
    }
    const table = detail.tables?.[tableIndex];
    if (rateView === 'matrix' && !tableHasDuration(table)) {
      setRateView('list');
    }
  }, [detail, tableIndex, rateView]);

  const closeModal = () => {
    setSelected(null);
  };

  const activeTable = detail?.tables?.[tableIndex];
  const canUseMatrix = Boolean(activeTable && tableHasDuration(activeTable));

  return (
    <div class="app-root">
      <header class="page-header">
        <div class="brand">
          <span class="logo">MORT</span>
          <div class="headline">
            <h1>Mortality Tables</h1>
            <p class="meta-inline">
              Search and inspect converted XTbML data • {filtered.length} tables
            </p>
          </div>
        </div>
        <label class="search-field">
          <span>Search</span>
          <input
            type="search"
            placeholder="Table name, provider, keywords…"
            value={query}
            onInput={(event) => setQuery((event.target as HTMLInputElement).value)}
          />
        </label>
      </header>

      <section class="table-panel">
        <div class="selection-bar">
          <span>{selectedCount} selected</span>
          <div class="selection-actions">
            <button
              onClick={() => handleBulkDownload('json')}
              disabled={selectedCount === 0 || bulkLoading !== null}
            >
              {bulkLoading === 'json' ? 'JSON Zip…' : 'JSON Zip'}
            </button>
            <button
              onClick={() => handleBulkDownload('csv')}
              disabled={selectedCount === 0 || bulkLoading !== null}
            >
              {bulkLoading === 'csv' ? 'CSV Zip…' : 'CSV Zip'}
            </button>
            <button onClick={clearSelection} disabled={selectedCount === 0}>
              Clear
            </button>
          </div>
        </div>

        <div class="scroll-frame" ref={listRef}>
          <table class="results-table">
            <thead>
              <tr>
                <th class="select-col">
                  <input
                    ref={selectAllRef}
                      type="checkbox"
                    onChange={handleSelectAllFiltered}
                    aria-label="Select all filtered tables"
                  />
                </th>
                <th>ID</th>
                <th>Name</th>
                <th class="download-col">Download</th>
              </tr>
            </thead>
            <tbody>
              {visibleTables.map((table) => (
                <tr key={table.detailPath} onClick={() => setSelected(table)}>
                  <td class="select-col">
                    <input
                      type="checkbox"
                      checked={selectedKeys.has(table.detailPath)}
                      onClick={(event) => {
                        event.stopPropagation();
                        const target = event.currentTarget as HTMLInputElement;
                        handleCheckboxSelection(table.detailPath, {
                          shiftKey: event.shiftKey,
                          nextChecked: target.checked,
                        });
                      }}
                      aria-label={`Select ${table.name}`}
                    />
                  </td>
                  <td>{table.tableIdentity}</td>
                  <td>
                    <span class="name">{table.name}</span>
                    <span class="summary">{table.summary}</span>
                  </td>
                  <td class="download-col">
                    <div class="download-buttons" onClick={(event) => event.stopPropagation()}>
                      <button onClick={() => triggerJsonDownload(table.detailPath, table.identifier)}>JSON</button>
                      <button
                        onClick={() => handleSummaryCsvDownload(table)}
                        disabled={csvDownloadingId === table.detailPath}
                      >
                        {csvDownloadingId === table.detailPath ? 'CSV…' : 'CSV'}
                      </button>
                    </div>
                  </td>
                </tr>
              ))}
                {visibleTables.length === 0 && (
                  <tr>
                    <td colSpan={4} class="empty">
                      No tables match “{query}”.
                    </td>
                  </tr>
                )}
            </tbody>
          </table>
        </div>
      </section>

      {selected && (
        <div class="modal-overlay" onClick={closeModal}>
          <div class="modal" role="dialog" aria-modal="true" onClick={(event) => event.stopPropagation()}>
            <header class="modal-header">
              <div>
                <p class="eyebrow">Table {selected.tableIdentity}</p>
                <h2>{selected.name}</h2>
                <p class="provider">{detail?.classification?.providerName ?? selected.provider}</p>
              </div>
              <button class="close-btn" onClick={closeModal}>
                Close
              </button>
            </header>

            {detailState === 'loading' && <p class="status">Loading detail…</p>}
            {detailState === 'error' && <p class="status error">Unable to load detail.</p>}
            {detail && (
              <>
                <nav class="tab-bar">
                  {TAB_ORDER.map((tab) => (
                    <button
                      key={tab}
                      class={tab === activeTab ? 'active' : ''}
                      onClick={() => setActiveTab(tab)}
                    >
                      {tabLabel(tab)}
                    </button>
                  ))}
                </nav>

                <div class="tab-content">
                  {activeTab === 'classification' && renderClassification(detail)}
                  {activeTab === 'metadata' && renderMetadata(detail, tableIndex)}
                  {activeTab === 'rates' && (
                    <>
                      <div class="rate-view-toggle">
                        <span>Rates View</span>
                        <div class="toggle-buttons">
                          <button
                            class={rateView === 'list' ? 'active' : ''}
                            onClick={() => setRateView('list')}
                            aria-pressed={rateView === 'list'}
                          >
                            List
                          </button>
                          <button
                            class={rateView === 'matrix' ? 'active' : ''}
                            onClick={() => setRateView('matrix')}
                            disabled={!canUseMatrix}
                            aria-pressed={rateView === 'matrix'}
                          >
                            Matrix
                          </button>
                        </div>
                      </div>
                      {renderRates(detail, tableIndex, rateView)}
                    </>
                  )}
                </div>

                <footer class="modal-footer">
                  <div class="table-switcher">
                    <button
                      onClick={() => setTableIndex((idx) => Math.max(0, idx - 1))}
                      disabled={!detail.tables || tableIndex === 0}
                    >
                      Prev Table
                    </button>
                    <span>
                      {detail.tables && detail.tables.length > 0
                        ? `Table ${tableIndex + 1} of ${detail.tables.length}`
                        : 'No rate tables'}
                    </span>
                    <button
                      onClick={() =>
                        setTableIndex((idx) => Math.min((detail.tables?.length ?? 1) - 1, idx + 1))
                      }
                      disabled={!detail.tables || tableIndex >= (detail.tables?.length ?? 0) - 1}
                    >
                      Next Table
                    </button>
                  </div>
                  <div class="download-actions">
                    <button
                      onClick={() => selected?.detailPath && triggerJsonDownload(selected.detailPath, detail?.identifier)}
                      disabled={!selected?.detailPath}
                    >
                      JSON
                    </button>
                    <button onClick={handleDetailCsvDownload} disabled={!detail?.tables?.[tableIndex]?.rates?.length}>
                      CSV
                    </button>
                  </div>
                </footer>
              </>
            )}
          </div>
        </div>
      )}
    </div>
  );
}

function tabLabel(tab: DetailTab): string {
  switch (tab) {
    case 'classification':
      return 'Classification';
    case 'metadata':
      return 'Metadata';
    case 'rates':
      return 'Rates';
  }
}

function renderClassification(detail: ConvertedTable) {
  const c = detail.classification;
  if (!c) {
    return <p>No classification data.</p>;
  }
  return (
    <dl class="kv">
      <dt>Table Identity</dt>
      <dd>{c.tableIdentity}</dd>
      <dt>Provider</dt>
      <dd>{c.providerName ?? 'Unknown'}</dd>
      <dt>Domain</dt>
      <dd>{c.providerDomain ?? '—'}</dd>
      <dt>Reference</dt>
      <dd>{c.tableReference ?? '—'}</dd>
      <dt>Content Type</dt>
      <dd>{c.contentType ? `${c.contentType.label} (${c.contentType.code})` : '—'}</dd>
      <dt>Description</dt>
      <dd>{c.tableDescription ?? '—'}</dd>
      {c.comments && (
        <>
          <dt>Comments</dt>
          <dd>{c.comments}</dd>
        </>
      )}
      {c.keywords && c.keywords.length > 0 && (
        <>
          <dt>Keywords</dt>
          <dd>{c.keywords.join(', ')}</dd>
        </>
      )}
    </dl>
  );
}

function renderMetadata(detail: ConvertedTable, tableIndex: number) {
  const table = detail.tables?.[tableIndex];
  if (!table || !table.metadata) {
    return <p>No metadata attached to this table.</p>;
  }
  const meta = table.metadata;
  return (
    <dl class="kv">
      <dt>Scaling Factor</dt>
      <dd>{meta.scalingFactor ?? '—'}</dd>
      <dt>Data Type</dt>
      <dd>
        {meta.dataType ? `${meta.dataType.label} (${meta.dataType.code})` : '—'}
      </dd>
      <dt>Nation</dt>
      <dd>
        {meta.nation ? `${meta.nation.label} (${meta.nation.code})` : '—'}
      </dd>
      <dt>Description</dt>
      <dd>{meta.tableDescription ?? '—'}</dd>
      {(meta.axes?.length ?? 0) > 0 && (
        <>
          <dt>Axes</dt>
          <dd>
            <ul>
              {meta.axes?.map((axis) => (
                <li key={axis.id}>
                  <strong>{axis.axisName}</strong> ({axis.id}) — {axis.minValue} to {axis.maxValue} step{' '}
                  {axis.increment} ({axis.scaleType.label})
                </li>
              ))}
            </ul>
          </dd>
        </>
      )}
    </dl>
  );
}

function renderRates(detail: ConvertedTable, tableIndex: number, viewMode: RateView) {
  const table = detail.tables?.[tableIndex];
  if (!table || !table.rates || table.rates.length === 0) {
    return <p>No rates found for this table.</p>;
  }
  if (viewMode === 'matrix') {
    const matrix = buildRateMatrix(table);
    if (!matrix) {
      return (
        <>
          <p>Matrix view is available only when durations are present.</p>
          {renderRateList(table)}
        </>
      );
    }
    return renderRateMatrix(matrix);
  }
  return renderRateList(table);
}

function renderRateList(table: TablePayload) {
  const rates = table.rates ?? [];
  if (rates.length === 0) {
    return <p>No rates found for this table.</p>;
  }
  const hasDuration = tableHasDuration(table);
  return (
    <div class="rates-scroll list-scroll">
      <table class="list-table">
        <thead>
          <tr>
            <th>Age</th>
            {hasDuration && <th>Duration</th>}
            <th>Rate</th>
          </tr>
        </thead>
        <tbody>
          {rates.map((rate, idx) => (
            <tr key={`${rate.age}-${idx}`}>
              <td>{rate.age}</td>
              {hasDuration && <td>{rate.duration ?? '—'}</td>}
              <td>{typeof rate.rate === 'number' ? rate.rate.toFixed(6) : '—'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

type RateMatrix = {
  ages: number[];
  durations: number[];
  values: Map<number, Map<number, number | null>>;
};

function buildRateMatrix(table: TablePayload): RateMatrix | null {
  const rates = table.rates ?? [];
  const durations = Array.from(
    new Set(
      rates
        .map((rate) => (typeof rate.duration === 'number' ? rate.duration : null))
        .filter((dur): dur is number => typeof dur === 'number')
    )
  ).sort((a, b) => a - b);
  if (durations.length === 0) {
    return null;
  }
  const ages = Array.from(new Set(rates.map((rate) => rate.age))).sort((a, b) => a - b);
  const values = new Map<number, Map<number, number | null>>();
  for (const rate of rates) {
    if (typeof rate.duration !== 'number') {
      continue;
    }
    const row = values.get(rate.age) ?? new Map<number, number | null>();
    row.set(rate.duration, typeof rate.rate === 'number' ? rate.rate : null);
    values.set(rate.age, row);
  }
  return { ages, durations, values };
}

function renderRateMatrix(matrix: RateMatrix) {
  return (
    <div class="rates-scroll matrix-scroll">
      <table class="matrix-table">
        <thead>
          <tr>
            <th>Age</th>
            {matrix.durations.map((dur) => (
              <th key={dur}>Dur {dur}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {matrix.ages.map((age) => (
            <tr key={age}>
              <td>{age}</td>
              {matrix.durations.map((dur) => {
                const value = matrix.values.get(age)?.get(dur);
                return (
                  <td key={dur}>{typeof value === 'number' ? value.toFixed(6) : '—'}</td>
                );
              })}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function tableHasDuration(table?: TablePayload | null): boolean {
  if (!table?.rates) {
    return false;
  }
  return table.rates.some((rate) => typeof rate.duration === 'number');
}

function buildCsv(detail: ConvertedTable, tableIndex: number): string {
  const table = detail.tables?.[tableIndex];
  if (!table || !table.rates || table.rates.length === 0) {
    return '';
  }
  const header = ['tableIndex', 'age', 'duration', 'rate'];
  const rows = table.rates.map((rate) => [
    table.index ?? tableIndex,
    rate.age,
    rate.duration ?? '',
    typeof rate.rate === 'number' ? rate.rate : '',
  ]);
  const versionNote = detail.version ? `# Version: ${detail.version}\n` : '';
  const identifierNote = detail.identifier ? `# Identifier: ${detail.identifier}\n` : '';
  const csvBody = [header, ...rows]
    .map((row) => row.map((value) => `"${value ?? ''}"`).join(','))
    .join('\n');
  return `${identifierNote}${versionNote}${csvBody}\n`;
}
