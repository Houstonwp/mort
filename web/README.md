# Mortality Tables Web UI

This Astro + Preact front‑end mirrors the existing TUI experience with a clean, modern shell:

- Instant fuzzy search (Fuse.js) across table identifiers, names, providers, and keywords.
- Server-rendered pagination with a sortable list that matches the TUI ordering.
- A modal detail view with the same three tabs as the TUI: Classification, Metadata, and Rates. Tabs load lazily from static JSON detail files generated at build time.
- Lightweight styling focused on readability and performance—no heavy component frameworks.

All data comes from the repository’s `json/` directory. During `astro build` we hydrate an index for the list and pre-render `/detail/<identifier>.json` endpoints for modal fetches, so the UI stays fully static after deployment.

## Project Structure

```
web/
├── src/
│   ├── components/MortalityApp.tsx     # Main island with search, table, modal
│   ├── lib/loadTables.ts               # Reads and sorts JSON summaries
│   ├── lib/types.ts                    # Shared TypeScript contracts
│   ├── pages/index.astro               # Shell page that loads the island
│   └── pages/detail/[identifier].json  # Pre-rendered detail endpoints
└── src/styles/app.css                  # Minimal design system
```

## Commands

Run everything from `web/`:

| Command           | Purpose                                                     |
| ----------------- | ----------------------------------------------------------- |
| `npm install`     | Install Astro, Preact, Fuse.js, and supporting deps         |
| `npm run dev`     | Start the local dev server at `http://localhost:4321`       |
| `npm run build`   | Build the static site + JSON detail payloads into `dist/`   |
| `npm run preview` | Preview the production build                                |

> `npm run build` requires the repository-level `json/` directory to be present because it reads those files to create the index and detail payloads.

## Notes

- If you add or remove mortality JSON files, simply rerun `npm run build` to update the static detail endpoints.
- Styling favors pure CSS for fast loading; feel free to layer in tokens or themes as the design system evolves.
