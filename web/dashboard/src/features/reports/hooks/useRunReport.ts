import { useEffect, useState } from "react";
import { errorMessage, initialResource } from "../../../shared/utils";
import type { Resource } from "../../../shared/utils";
import type { RunHistoryItem } from "../../runs";
import { getJSONReport } from "../api";
import type { RunReport } from "../types";

export function useRunReport(run: RunHistoryItem | null): Resource<RunReport> {
  const [report, setReport] = useState<Resource<RunReport>>(initialResource(false));
  const runId = run?.runId;

  useEffect(() => {
    let cancelled = false;

    if (!runId) {
      setReport({ data: null, error: null, loading: false });
      return () => {
        cancelled = true;
      };
    }

    setReport({ data: null, error: null, loading: true });
    void getJSONReport(runId)
      .then((data) => {
        if (!cancelled) {
          setReport({ data, error: null, loading: false });
        }
      })
      .catch((error) => {
        if (!cancelled) {
          setReport({ data: null, error: errorMessage(error), loading: false });
        }
      });

    return () => {
      cancelled = true;
    };
  }, [runId]);

  return report;
}
