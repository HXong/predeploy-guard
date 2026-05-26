# PreDeploy Guard Dashboard

Local React + TypeScript dashboard for the PreDeploy Guard API.

## Backend

From the repository root:

```bash
go run ./cmd/predeploy serve --config examples/predeploy.yaml
```

The dashboard dev server proxies `/api` requests to `http://localhost:7070`.

## Install

```bash
npm install
```

## Develop

```bash
npm run dev
```

## Build

```bash
npm run build
```

Set `VITE_PREDEPLOY_API_BASE_URL` to override the API base URL when needed. By default, the frontend uses relative `/api` requests.
