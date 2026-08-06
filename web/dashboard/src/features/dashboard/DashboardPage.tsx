import { ConfigSummaryCard, ExecutionPlanCard } from "../config";
import { RunDetailsPanel } from "../reports";
import { RunHistoryTable, RunTasksTable, TaskLogsPanel, TriggerRunPanel } from "../runs";
import { HealthCard } from "./components/HealthCard";
import { useDashboardController } from "./hooks/useDashboardController";

export function DashboardPage() {
  const dashboard = useDashboardController();

  return (
    <>
      <section className="panel-actions dashboard-actions">
        <button className="secondary-button" type="button" onClick={dashboard.refreshAll}>
          Refresh
        </button>
      </section>

      <section className="top-grid">
        <HealthCard
          data={dashboard.health.data}
          loading={dashboard.health.loading}
          error={dashboard.health.error}
        />
        <ConfigSummaryCard
          data={dashboard.summary.data}
          loading={dashboard.summary.loading}
          error={dashboard.summary.error}
        />
        <TriggerRunPanel
          error={dashboard.runTasks.triggerError}
          onProfileChange={dashboard.runTasks.setProfile}
          onTriggerRun={dashboard.handleTriggerRun}
          profile={dashboard.runTasks.profile}
          triggering={dashboard.runTasks.triggering}
        />
      </section>

      <section className="content-grid">
        <RunTasksTable
          error={dashboard.runTasks.tasks.error}
          loading={dashboard.runTasks.tasks.loading}
          onSelect={dashboard.runTasks.setSelectedTaskId}
          selectedTaskId={dashboard.runTasks.selectedTaskId}
          tasks={dashboard.runTasks.tasks.data ?? []}
        />
        <TaskLogsPanel
          error={dashboard.taskLogs.taskLogs.error}
          loading={dashboard.taskLogs.taskLogs.loading}
          logs={dashboard.taskLogs.taskLogs.data?.logs ?? []}
          task={dashboard.taskLogs.selectedTask.data}
          taskError={dashboard.taskLogs.selectedTask.error}
        />
      </section>

      <section className="content-grid">
        <RunHistoryTable
          error={dashboard.history.error}
          loading={dashboard.history.loading}
          onSelect={dashboard.setSelectedRunId}
          runs={dashboard.latestHistory}
          selectedRunId={dashboard.selectedRunId}
        />
        <RunDetailsPanel
          markdownPreview={dashboard.markdownPreview.markdownPreview}
          onPreviewMarkdown={dashboard.markdownPreview.handlePreviewMarkdown}
          report={dashboard.runReport}
          run={dashboard.selectedRun}
        />
      </section>

      <section className="content-grid single-grid">
        <ExecutionPlanCard
          data={dashboard.explain.data}
          loading={dashboard.explain.loading}
          error={dashboard.explain.error}
        />
      </section>
    </>
  );
}
