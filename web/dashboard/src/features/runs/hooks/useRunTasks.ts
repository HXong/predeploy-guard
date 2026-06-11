import { useCallback, useState } from "react";
import { errorMessage, initialResource } from "../../../shared/utils";
import type { Resource } from "../../../shared/utils";
import { getRunTasks, triggerRun } from "../api";
import type { RunTask } from "../types";

type TriggerOptions = {
  onTaskStarted: (task: RunTask) => void;
};

export function useRunTasks() {
  const [tasks, setTasks] = useState<Resource<RunTask[]>>(initialResource);
  const [selectedTaskId, setSelectedTaskId] = useState("");
  const [profile, setProfile] = useState("smoke-only");
  const [triggering, setTriggering] = useState(false);
  const [triggerError, setTriggerError] = useState<string | null>(null);

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

  const handleTriggerRun = useCallback(
    async ({ onTaskStarted }: TriggerOptions) => {
      setTriggering(true);
      setTriggerError(null);

      try {
        const task = await triggerRun(profile);
        onTaskStarted(task);
        setSelectedTaskId(task.taskId);
        await loadTasks();
      } catch (error) {
        setTriggerError(errorMessage(error));
      } finally {
        setTriggering(false);
      }
    },
    [loadTasks, profile],
  );

  return {
    handleTriggerRun,
    loadTasks,
    profile,
    selectedTaskId,
    setProfile,
    setSelectedTaskId,
    tasks,
    triggerError,
    triggering,
  };
}
