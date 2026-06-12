import { useMemo, useState } from "react";
import { Card } from "../../shared/components";
import { defaultConfigBuilderState } from "./defaults";
import { generatePredeployYaml } from "./yaml";
import { validateConfigBuilder } from "./validation";
import { DependencyBuilder } from "./components/DependencyBuilder";
import { ServiceForm } from "./components/ServiceForm";
import { YamlPreview } from "./components/YamlPreview";

export function ConfigBuilderPage() {
  const [config, setConfig] = useState(defaultConfigBuilderState);
  const yaml = useMemo(() => generatePredeployYaml(config), [config]);
  const errors = useMemo(() => validateConfigBuilder(config), [config]);

  return (
    <section className="content-grid builder-grid">
      <div className="builder-stack">
        <Card title="Service Configuration">
          <ServiceForm config={config} onChange={setConfig} />
        </Card>
        <Card title="Dependencies">
          <DependencyBuilder config={config} onChange={setConfig} />
        </Card>
      </div>
      <Card title="predeploy.yaml Preview">
        <YamlPreview errors={errors} yaml={yaml} />
      </Card>
    </section>
  );
}
