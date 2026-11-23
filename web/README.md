# Web Dashboard

The **Web Dashboard** (`web`) is a Vue 3.5+ single-page application for launching website crawls, browsing the 435 Audubon species catalog, and inspecting bird sightings served by `services/catalog`.

## Stack

- **Framework & Routing**: Vue 3.5 (`<script setup lang="ts">`) + Vue Router 5
- **Server State**: TanStack Vue Query v5 (`@tanstack/vue-query`)
- **Styling**: Tailwind CSS v4 (`@tailwindcss/vite`)
- **Tooling**: Vite 8, `vue-tsc` powered by `typescript-go` (`typescript-native-bridge`), Oxlint, ESLint, and Oxfmt

---

## Prerequisites & IDE Setup

- **Node.js** `26+` (LTS, pinned in `.node-version`)
- **pnpm** `10+` — Node 26 no longer bundles Corepack, so install `pnpm` first:

```bash
npm install -g pnpm
# or standalone installer:
# curl -fsSL https://get.pnpm.io/install.sh | sh -
```

Install the recommended VS Code extensions (listed in `.vscode/extensions.json`):

- **Vue (Official)** (`Vue.volar`) — disables/replaces legacy Vetur; required for `.vue` TypeScript support
- **Tailwind CSS IntelliSense** (`bradlc.vscode-tailwindcss`) — Tailwind v4 class completion and `@theme` support
- **Oxc** (`oxc.oxc-vscode`) — Oxlint diagnostics and Oxfmt formatting on save
- **ESLint** (`dbaeumer.vscode-eslint`) — Vue SFC template lint rules

After running `pnpm install`, accept the VS Code prompt to use the workspace TypeScript version (or run **TypeScript: Select TypeScript Version** -> **Use Workspace Version**) so editor checking uses `typescript-native-bridge` (`tsgo`).

---

## Configuration

Development defaults are committed in `.env.development`. To override locally (e.g. pointing to a remote media CDN), copy `.env.example` to `.env.local`:

```bash
cp .env.example .env.local
```

| Variable | Default | Description |
| :--- | :--- | :--- |
| `VITE_MEDIA_BASE_URL` | `http://localhost:8081` | Base URL for Audubon plate images and bird call audio (served locally by the `media` Docker container). |

In development, Vite proxies `/api/*` requests directly to the local `catalog` service at `http://localhost:8080`.

---

## Development

Install dependencies:

```bash
pnpm install
```

Start the dev server (`http://localhost:5173`):

```bash
pnpm dev
```

Type-check, lint, format, and build (or run `make lint-web` / `make format-web` from the repo root):

```bash
pnpm type-check
pnpm lint
pnpm format
pnpm build
```
