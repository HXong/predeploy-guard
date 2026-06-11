import type { Dispatch, SetStateAction } from "react";
import { useCallback, useEffect, useMemo, useState } from "react";
import { getConfigExplain, getConfigSummary } from "../../config";
import type { ConfigExplanation, ConfigSummary } from "../../config";
import { getHistory, useRunTasks, useTaskLogs } from "../../runs";
import type { RunHistoryItem, RunTask } from "../../runs";
import { useMarkdownPreview } from "../../reports";
import { errorMessage, initialResource } from "../../../shared/utils";
import type { Resource } from "../../../shared/utils";
import { getHealth } from "../api";
import type { HealthResponse } from "../types";

const runningStatuses = new Set<RunTask["status"]>(["queued", "running"]);

export function useDashboardController() {
  const [health, setHealth] = useState<Resource<HealthResponse>>(initialResource);
  const [summary, setSummary] = useState<Resource<ConfigSummary>>(initialResource);
  const [explain, setExplain] = useState<Resource<ConfigExplanation>>(initialResource);
  const [history, setHistory] = useState<Resource<RunHistoryItem[]>>(initialResource);
  const [selectedRunId, setSelectedRunId] = useState("");

  const runTasks = useRunTasks();
  const taskLogs = useTaskLogs();
  const { loadSelectedTask, selectedTask, setTriggeredTask } = taskLogs;
  const { handleTriggerRun: triggerRunTask, loadTasks, selectedTaskId } = runTasks;

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
  const markdownPreview = useMarkdownPreview(selectedRun);

  const handleTriggerRun = useCallback(() => {
    void triggerRunTask({ onTaskStarted: setTriggeredTask });
  }, [setTriggeredTask, triggerRunTask]);

  const refreshAll = useCallback(() => {
    loadHealth();
    loadSummary();
    loadExplain();
    loadHistory();
    void loadTasks();
    if (selectedTaskId) {
      void loadSelectedTask(selectedTaskId);
    }
  }, [loadExplain, loadHealth, loadHistory, loadSelectedTask, loadSummary, loadTasks, selectedTaskId]);

  return {
    explain,
    health,
    history,
    latestHistory,
    markdownPreview,
    refreshAll,
    runTasks,
    selectedRun,
    selectedRunId,
    setSelectedRunId,
    summary,
    taskLogs,
    handleTriggerRun,
  };
}

function loadResource<T>(setter: Dispatch<SetStateAction<Resource<T>>>, loader: () => Promise<T>) {
  setter((current) => ({ ...current, loading: true, error: null }));
  void loader()
    .then((data) => setter({ data, error: null, loading: false }))
    .catch((error) => setter({ data: null, error: errorMessage(error), loading: false }));
}
