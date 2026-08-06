import { useCallback, useEffect, useState } from "react";
import { errorMessage, initialResource } from "../../../shared/utils";
import type { Resource } from "../../../shared/utils";
import type { RunHistoryItem } from "../../runs";
import { getMarkdownReport } from "../api";

export function useMarkdownPreview(run: RunHistoryItem | null) {
  const [markdownPreview, setMarkdownPreview] = useState<Resource<string>>(initialResource(false));

  useEffect(() => {
    setMarkdownPreview({ data: null, error: null, loading: false });
  }, [run?.runId]);

  const handlePreviewMarkdown = useCallback(async () => {
    if (!run) {
      return;
    }

    setMarkdownPreview({ data: null, error: null, loading: true });

    try {
      const markdown = await getMarkdownReport(run.runId);
      setMarkdownPreview({ data: markdown, error: null, loading: false });
    } catch (error) {
      setMarkdownPreview({ data: null, error: errorMessage(error), loading: false });
    }
  }, [run]);

  return {
    handlePreviewMarkdown,
    markdownPreview,
  };
}
