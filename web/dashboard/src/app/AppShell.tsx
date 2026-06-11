import { useState } from "react";
import { ConfigBuilderPage } from "../features/config-builder";
import { DashboardPage } from "../features/dashboard";
import { Header } from "../features/dashboard/components/Header";

export type AppView = "dashboard" | "config-builder";

export function AppShell() {
  const [activeView, setActiveView] = useState<AppView>("dashboard");

  return (
    <main className="app-shell">
      <Header />

      <nav className="view-tabs" aria-label="Dashboard views">
        <button
          className={activeView === "dashboard" ? "chip active" : "chip"}
          type="button"
          onClick={() => setActiveView("dashboard")}
        >
          Dashboard
        </button>
        <button
          className={activeView === "config-builder" ? "chip active" : "chip"}
          type="button"
          onClick={() => setActiveView("config-builder")}
        >
          Config Builder
        </button>
      </nav>

      {activeView === "dashboard" ? <DashboardPage /> : <ConfigBuilderPage />}
    </main>
  );
}
