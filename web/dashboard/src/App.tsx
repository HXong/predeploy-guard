import { useCallback, useEffect, useMemo, useState } from "react";
import type { Dispatch, ReactNode, SetStateAction } from "react";
import {
  ConfigExplanation,
  ConfigSummary,
  getConfigExplain,
  getConfigSummary,
  getHealth,
  getHistory,
  getRunTask,
  getRunTaskLogs,
  getRunTasks,
  HealthResponse,
  RunHistoryItem,
  RunTask,
  RunTaskLogsResponse,
  triggerRun,
} from "./api/client";

type Resource<T> = {
  data: T | null;
  error: string | null;
  loading: boolean;
};

const profiles = ["base", "smoke-only", "light-load", "stress-test"];

function initialResource<T>(): Resource<T> {
  return {
    data: null,
    error: null,
    loading: true,
  };
}

export default function App() {
  const [health, setHealth] = useState<Resource<HealthResponse>>(initialResource);
  const [summary, setSummary] = useState<Resource<ConfigSummary>>(initialResource);
  const [explain, setExplain] = useState<Resource<ConfigExplanation>>(initialResource);
  const [history, setHistory] = useState<Resource<RunHistoryItem[]>>(initialResource);
  const [tasks, setTasks] = useState<Resource<RunTask[]>>(initialResource);
  const [selectedTaskId, setSelectedTaskId] = useState<string>("");
  const [selectedTask, setSelectedTask] = useState<Resource<RunTask>>({
    data: null,
    error: null,
    loading: false,
  });
  const [taskLogs, setTaskLogs] = useState<Resource<RunTaskLogsResponse>>({
    data: null,
    error: null,
    loading: false,
  });
  const [profile, setProfile] = useState("smoke-only");
  const [triggering, setTriggering] = useState(false);
  const [triggerError, setTriggerError] = useState<string | null>(null);

  const loadHealth = useCallback(() => {
    loadResource(setHealth, getHealth);
  }, []);

  const loadSummary = useCallback(() => {
    loadResource(setSummary, getConfigSummary);
  }, []);

  const loadExplain = useCallback(() => {
    loadResource(setExplain, getConfigExplain);
  }, []);

  const loadHistory = useCallback(() => {
    loadResource(setHistory, getHistory);
  }, []);

  const loadTasks = useCallback(async () => {
    setTasks((current) => ({ ...current, loading: true, error: null }));

    try {
      const data = await getRunTasks();
      setTasks({ data, error: null, loading: false });
      setSelectedTaskId((current) => current || data[0]?.taskId || "");
    } catch (error) {
      setTasks({ data: null, error: errorMessage(error), loading: false });
    }
  }, []);

  const loadSelectedTask = useCallback(async (taskId: string) => {
    if (!taskId) {
      setSelectedTask({ data: null, error: null, loading: false });
      setTaskLogs({ data: null, error: null, loading: false });
      return;
    }

    setSelectedTask((current) => ({ ...current, loading: true, error: null }));
    setTaskLogs((current) => ({ ...current, loading: true, error: null }));

    const [taskResult, logsResult] = await Promise.allSettled([getRunTask(taskId), getRunTaskLogs(taskId)]);

    if (taskResult.status === "fulfilled") {
      setSelectedTask({ data: taskResult.value, error: null, loading: false });
    } else {
      setSelectedTask({ data: null, error: errorMessage(taskResult.reason), loading: false });
    }

    if (logsResult.status === "fulfilled") {
      setTaskLogs({ data: logsResult.value, error: null, loading: false });
    } else {
      setTaskLogs({ data: null, error: errorMessage(logsResult.reason), loading: false });
    }
  }, []);

  useEffect(() => {
    loadHealth();
    loadSummary();
    loadExplain();
    loadHistory();
    void loadTasks();
  }, [loadExplain, loadHealth, loadHistory, loadSummary, loadTasks]);

  useEffect(() => {
    void loadSelectedTask(selectedTaskId);
  }, [loadSelectedTask, selectedTaskId]);

  useEffect(() => {
    const task = selectedTask.data;
    if (!task || (task.status !== "queued" && task.status !== "running")) {
      return;
    }

    const intervalId = window.setInterval(() => {
      void loadTasks();
      void loadSelectedTask(task.taskId);
    }, 3000);

    return () => window.clearInterval(intervalId);
  }, [loadSelectedTask, loadTasks, selectedTask.data]);

  const latestHistory = useMemo(() => history.data?.slice().reverse().slice(0, 8) ?? [], [history.data]);

  async function handleTriggerRun() {
    setTriggering(true);
    setTriggerError(null);

    try {
      const task = await triggerRun(profile);
      setSelectedTaskId(task.taskId);
      await loadTasks();
    } catch (error) {
      setTriggerError(errorMessage(error));
    } finally {
      setTriggering(false);
    }
  }

  return (
    <main className="app-shell">
      <header className="page-header">
        <div>
          <p className="eyebrow">Local validation control plane</p>
          <h1>PreDeploy Guard Dashboard</h1>
        </div>
        <button className="secondary-button" type="button" onClick={refreshAll}>
          Refresh
        </button>
      </header>

      <section className="top-grid">
        <Card title="API Health" loading={health.loading} error={health.error}>
          {health.data && (
            <div className="status-row">
              <span className={`status-dot ${health.data.status === "ok" ? "ok" : "warn"}`} />
              <span className="status-text">{health.data.status}</span>
            </div>
          )}
        </Card>

        <Card title="Config Summary" loading={summary.loading} error={summary.error}>
          {summary.data && (
            <div className="summary-grid">
              <Metric label="Service" value={summary.data.service.name} />
              <Metric label="Runtime" value={summary.data.runtime.type} />
              <Metric label="Port" value={String(summary.data.service.port)} />
              <Metric label="Profiles" value={String(summary.data.counts.profiles)} />
              <Metric label="Smoke checks" value={String(summary.data.counts.smokeChecks)} />
              <Metric label="Dependencies" value={String(summary.data.counts.dependencies)} />
            </div>
          )}
        </Card>

        <Card title="Trigger Run">
          <div className="profile-panel">

            <div className="label-container">
              <label htmlFor="profile">Profile (optional)</label>
              <span 
                className="helper-icon" 
                title="Select or type a configuration profile to run. Leave empty to run base config."
              >
                (?)
              </span>
            </div>

            <input
              id="profile"
              value={profile}
              onChange={(event) => setProfile(event.target.value)}
              placeholder="base"
            />
            <div className="profile-buttons">
              {profiles.map((candidate) => (
                <button
                  className={candidate === profile ? "chip active" : "chip"}
                  key={candidate}
                  type="button"
                  onClick={() => setProfile(candidate)}
                >
                  {candidate}
                </button>
              ))}
            </div>
            <button className="primary-button" type="button" disabled={triggering} onClick={handleTriggerRun}>
              {triggering ? "Starting..." : "Start validation"}
            </button>
            {triggerError && <p className="error-text">{triggerError}</p>}
          </div>
        </Card>
      </section>

      <section className="content-grid">
        <Card title="Run Tasks" loading={tasks.loading} error={tasks.error}>
          <TaskList tasks={tasks.data ?? []} selectedTaskId={selectedTaskId} onSelect={setSelectedTaskId} />
        </Card>

        <Card title="Task Logs" loading={taskLogs.loading} error={taskLogs.error}>
          <TaskLogPanel task={selectedTask.data} logs={taskLogs.data?.logs ?? []} taskError={selectedTask.error} />
        </Card>
      </section>

      <section className="content-grid">
        <Card title="Run History" loading={history.loading} error={history.error}>
          <HistoryTable runs={latestHistory} />
        </Card>

        <Card title={explain.data?.title ?? "Execution Plan"} loading={explain.loading} error={explain.error}>
          <div className="plan-list">
            {(explain.data?.steps.slice(0, 4) ?? []).map((step) => (
              <div className="plan-step" key={step.number}>
                <span>{step.number}</span>
                <div>
                  <strong>{step.title}</strong>
                  <p>{step.details[0]}</p>
                </div>
              </div>
            ))}
          </div>
        </Card>
      </section>
    </main>
  );

  function refreshAll() {
    loadHealth();
    loadSummary();
    loadExplain();
    loadHistory();
    void loadTasks();
    if (selectedTaskId) {
      void loadSelectedTask(selectedTaskId);
    }
  }
}

function Card({
  title,
  loading,
  error,
  children,
}: {
  title: string;
  loading?: boolean;
  error?: string | null;
  children: ReactNode;
}) {
  return (
    <article className="card">
      <div className="card-header">
        <h2>{title}</h2>
        {loading && <span className="loading-pill">Loading</span>}
      </div>
      {error ? <p className="empty-state">{error}</p> : children}
    </article>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="metric">
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}

function TaskList({
  tasks,
  selectedTaskId,
  onSelect,
}: {
  tasks: RunTask[];
  selectedTaskId: string;
  onSelect: (taskId: string) => void;
}) {
  if (tasks.length === 0) {
    return <p className="empty-state">No API-triggered runs yet.</p>;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Task</th>
            <th>Profile</th>
            <th>Status</th>
            <th>Created</th>
          </tr>
        </thead>
        <tbody>
          {tasks.map((task) => (
            <tr
              className={task.taskId === selectedTaskId ? "selected-row" : ""}
              key={task.taskId}
              onClick={() => onSelect(task.taskId)}
            >
              <td>{shortId(task.taskId)}</td>
              <td>{task.profile || "base"}</td>
              <td>
                <StatusBadge status={task.status} />
              </td>
              <td>{formatDate(task.createdAt)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function TaskLogPanel({
  task,
  logs,
  taskError,
}: {
  task: RunTask | null;
  logs: string[];
  taskError: string | null;
}) {
  if (taskError) {
    return <p className="empty-state">{taskError}</p>;
  }

  if (!task) {
    return <p className="empty-state">Select a task to inspect its logs.</p>;
  }

  return (
    <div className="logs-panel">
      <div className="task-summary">
        <StatusBadge status={task.status} />
        <span>{task.message || task.error || task.taskId}</span>
      </div>
      <pre>{logs.length > 0 ? logs.join("\n") : "No logs recorded yet."}</pre>
    </div>
  );
}

function HistoryTable({ runs }: { runs: RunHistoryItem[] }) {
  if (runs.length === 0) {
    return <p className="empty-state">No completed run history yet.</p>;
  }

  return (
    <div className="table-wrap">
      <table>
        <thead>
          <tr>
            <th>Run</th>
            <th>Service</th>
            <th>Profile</th>
            <th>Result</th>
            <th>Finished</th>
          </tr>
        </thead>
        <tbody>
          {runs.map((run) => (
            <tr key={run.runId}>
              <td title={run.runId}>{shortenRunId(run.runId)}</td>
              <td>{run.serviceName}</td>
              <td>{run.profile || "base"}</td>
              <td>{run.passed ? "Passed" : "Failed"}</td>
              <td>{formatDate(run.finishedAt)}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

function StatusBadge({ status }: { status: RunTask["status"] }) {
  return <span className={`status-badge ${status}`}>{status}</span>;
}

function loadResource<T>(setter: Dispatch<SetStateAction<Resource<T>>>, loader: () => Promise<T>) {
  setter((current) => ({ ...current, loading: true, error: null }));
  void loader()
    .then((data) => setter({ data, error: null, loading: false }))
    .catch((error) => setter({ data: null, error: errorMessage(error), loading: false }));
}

function errorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }

  return "Something went wrong while talking to the PreDeploy Guard API.";
}

function shortId(value: string): string {
  if (value.length <= 16) {
    return value;
  }

  return `${value.slice(0, 10)}...${value.slice(-4)}`;
}

function shortenRunId(runId: string): string {
  return runId.length > 12 ? `${runId.slice(0, 18)}...${runId.slice(-8)}` : runId;
}

function formatDate(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return new Intl.DateTimeFormat(undefined, {
    dateStyle: "medium",
    timeStyle: "short",
  }).format(date);
}
