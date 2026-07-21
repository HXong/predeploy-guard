import type { FormEvent } from "react";
import type { ConfigBuilderState, PerformanceConfigDraft } from "../types";
import { PerformanceConfigEditor } from "./PerformanceConfigEditor";

type PerformanceBuilderProps = {
  config: ConfigBuilderState;
  onChange: (config: ConfigBuilderState) => void;
};

export function PerformanceBuilder({ config, onChange }: PerformanceBuilderProps) {
  function updatePerformance(performance: PerformanceConfigDraft) {
    onChange({
      ...config,
      performance,
    });
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <form className="builder-form" onSubmit={handleSubmit}>
      <PerformanceConfigEditor
        onChange={updatePerformance}
        performance={config.performance}
        smokeCheckCount={config.checks.smoke.length}
      />
    </form>
  );
}
