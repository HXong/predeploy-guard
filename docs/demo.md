# PreDeploy Guard Demo

This walkthrough uses the repository sample service and the default Docker Compose runtime. Kubernetes support is optional and requires a developer-managed cluster, a working kubeconfig context, `kubectl`, and images accessible to that cluster.

## Prerequisites

- Go 1.22 or newer, or an installed `predeploy` binary
- Docker with Docker Compose
- Node.js and npm for the dashboard

## Run the demo

From the repository root:

```bash
predeploy validate examples/predeploy.yaml
predeploy run examples/predeploy.yaml
predeploy serve --config examples/predeploy.yaml --addr localhost:7070
```

Use `go run ./cmd/predeploy` in place of `predeploy` when the binary is not installed. Keep the API server running, then start the dashboard in another terminal:

```bash
cd web/dashboard
npm install
npm run dev
```

Open `http://localhost:5173`. Vite proxies `/api` to `http://localhost:7070`.

## Dashboard walkthrough

1. Confirm **API Health** is OK and review the compact **Config Summary**.
2. Select a completed entry in **Run History**.
3. Review **Runtime Environment** and **Run Timeline** in **Run Details**.
4. Continue through gateway, workload, and performance sections when the selected run recorded them.
5. Use the Markdown and JSON links for the complete report, or preview Markdown in the dashboard.

## Generated files and cleanup

Runs create `examples/reports/` next to the sample config. The directory, its history index, and report files are intentionally ignored by git and can be removed after the demo when no local history is needed. Stop the API server and Vite with `Ctrl+C` in their terminals; runtime-owned cleanup is performed by the validation run.

Do not commit generated reports. They can contain absolute paths, raw tool output, or environment-specific diagnostics.

## Screenshot checklist

Keep the stable filenames `assets/dashboard.png` and `assets/report.png`. Before replacing them, verify that the images show only generic sample data and do not include:

- secrets or environment values;
- absolute filesystem paths or application file contents;
- raw Docker, k6, or runtime diagnostics; or
- private project or account information.

Prefer one overview with the config summary and one selected-run view with Run History and a structured report section such as Run Timeline or Performance Result.
