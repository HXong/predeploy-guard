import { useCallback, useState } from "react";
import { errorMessage, initialResource } from "../../../shared/utils";
import type { Resource } from "../../../shared/utils";
import { getRunTask, getRunTaskLogs } from "../api";
import type { RunTask, RunTaskLogsResponse } from "../types";

export function useTaskLogs() {
  const [selectedTask, setSelectedTask] = useState<Resource<RunTask>>(initialResource(false));
  const [taskLogs, setTaskLogs] = useState<Resource<RunTaskLogsResponse>>(initialResource(false));

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

  const setTriggeredTask = useCallback((task: RunTask) => {
    setSelectedTask({ data: task, error: null, loading: false });
    setTaskLogs({ data: { taskId: task.taskId, logs: task.logs }, error: null, loading: false });
  }, []);

  return {
    loadSelectedTask,
    selectedTask,
    setTriggeredTask,
    taskLogs,
  };
}
