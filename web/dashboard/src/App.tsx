import { useCallback, useEffect, useMemo, useState } from "react";
import type { Dispatch, SetStateAction } from "react";
import {
  getConfigExplain,
  getConfigSummary,
  getHealth,
  getHistory,
  getMarkdownReport,
  getRunTask,
  getRunTaskLogs,
  getRunTasks,
  triggerRun,
} from "./api/client";
import { ConfigSummaryCard } from "./components/ConfigSummaryCard";
import { ExecutionPlanCard } from "./components/ExecutionPlanCard";
import { Header } from "./components/Header";
import { HealthCard } from "./components/HealthCard";
import { RunDetailsPanel } from "./components/RunDetailsPanel";
import { RunHistoryTable } from "./components/RunHistoryTable";
import { RunTasksTable } from "./components/RunTasksTable";
import { TaskLogsPanel } from "./components/TaskLogsPanel";
import { TriggerRunPanel } from "./components/TriggerRunPanel";
import type {
  ConfigExplanation,
  ConfigSummary,
  HealthResponse,
  RunHistoryItem,
  RunTask,
  RunTaskLogsResponse,
} from "./types/api";

type Resource<T> = {
  data: T | null;
  error: string | null;
  loading: boolean;
};

const runningStatuses = new Set<RunTask["status"]>(["queued", "running"]);

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
  const [selectedTaskId, setSelectedTaskId] = useState("");
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
  const [selectedRunId, setSelectedRunId] = useState("");
  const [markdownPreview, setMarkdownPreview] = useState<Resource<string>>({
    data: null,
    error: null,
    loading: false,
  });

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

  const loadSelectedTask = useCallback(async (taskId: string, showLoading = true): Promise<RunTask | null> => {
    if (!taskId) {
      setSelectedTask({ data: null, error: null, loading: false });
      setTaskLogs({ data: null, error: null, loading: false });
      return null;
    }

    if (showLoading) {
      setSelectedTask((current) => ({ ...current, loading: true, error: null }));
      setTaskLogs((current) => ({ ...current, loading: true, error: null }));
    }

    const [taskResult, logsResult] = await Promise.allSettled([getRunTask(taskId), getRunTaskLogs(taskId)]);
    let nextTask: RunTask | null = null;

    if (taskResult.status === "fulfilled") {
      nextTask = taskResult.value;
      setSelectedTask({ data: taskResult.value, error: null, loading: false });
    } else {
      setSelectedTask({ data: null, error: errorMessage(taskResult.reason), loading: false });
    }

    if (logsResult.status === "fulfilled") {
      setTaskLogs({ data: logsResult.value, error: null, loading: false });
    } else {
      setTaskLogs({ data: null, error: errorMessage(logsResult.reason), loading: false });
    }

    return nextTask;
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
    if (!task || !runningStatuses.has(task.status)) {
      return;
    }

    const intervalId = window.setInterval(() => {
      void loadSelectedTask(task.taskId, false).then((nextTask) => {
        if (!nextTask) {
          return;
        }

        if (nextTask.status !== task.status) {
          void loadTasks();
        }

        if (!runningStatuses.has(nextTask.status)) {
          void loadTasks();
          loadHistory();
        }
      });
    }, 2000);

    return () => window.clearInterval(intervalId);
  }, [loadHistory, loadSelectedTask, loadTasks, selectedTask.data]);

  const latestHistory = useMemo(() => history.data?.slice().reverse().slice(0, 8) ?? [], [history.data]);
  const selectedRun = useMemo(
    () => history.data?.find((run) => run.runId === selectedRunId) ?? null,
    [history.data, selectedRunId],
  );

  useEffect(() => {
    setMarkdownPreview({ data: null, error: null, loading: false });
  }, [selectedRunId]);

  async function handleTriggerRun() {
    setTriggering(true);
    setTriggerError(null);

    try {
      const task = await triggerRun(profile);
      setSelectedTask({ data: task, error: null, loading: false });
      setTaskLogs({ data: { taskId: task.taskId, logs: task.logs }, error: null, loading: false });
      setSelectedTaskId(task.taskId);
      await loadTasks();
    } catch (error) {
      setTriggerError(errorMessage(error));
    } finally {
      setTriggering(false);
    }
  }

  async function handlePreviewMarkdown() {
    if (!selectedRun) {
      return;
    }

    setMarkdownPreview({ data: null, error: null, loading: true });

    try {
      const markdown = await getMarkdownReport(selectedRun.runId);
      setMarkdownPreview({ data: markdown, error: null, loading: false });
    } catch (error) {
      setMarkdownPreview({ data: null, error: errorMessage(error), loading: false });
    }
  }

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

  return (
    <main className="app-shell">
      <Header onRefresh={refreshAll} />

      <section className="top-grid">
        <HealthCard data={health.data} loading={health.loading} error={health.error} />
        <ConfigSummaryCard data={summary.data} loading={summary.loading} error={summary.error} />
        <TriggerRunPanel
          error={triggerError}
          onProfileChange={setProfile}
          onTriggerRun={handleTriggerRun}
          profile={profile}
          triggering={triggering}
        />
      </section>

      <section className="content-grid">
        <RunTasksTable
          error={tasks.error}
          loading={tasks.loading}
          onSelect={setSelectedTaskId}
          selectedTaskId={selectedTaskId}
          tasks={tasks.data ?? []}
        />
        <TaskLogsPanel
          error={taskLogs.error}
          loading={taskLogs.loading}
          logs={taskLogs.data?.logs ?? []}
          task={selectedTask.data}
          taskError={selectedTask.error}
        />
      </section>

      <section className="content-grid">
        <RunHistoryTable
          error={history.error}
          loading={history.loading}
          onSelect={setSelectedRunId}
          runs={latestHistory}
          selectedRunId={selectedRunId}
        />
        <RunDetailsPanel
          markdownPreview={markdownPreview}
          onPreviewMarkdown={handlePreviewMarkdown}
          run={selectedRun}
        />
      </section>

      <section className="content-grid single-grid">
        <ExecutionPlanCard data={explain.data} loading={explain.loading} error={explain.error} />
      </section>
    </main>
  );
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
