import { useMemo, useState } from "react";
import { Card } from "../../shared/components";
import { defaultConfigBuilderState } from "./defaults";
import { generatePredeployYaml } from "./yaml";
import { validateConfigBuilder } from "./validation";
import { ServiceForm } from "./components/ServiceForm";
import { YamlPreview } from "./components/YamlPreview";

export function ConfigBuilderPage() {
  const [config, setConfig] = useState(defaultConfigBuilderState);
  const yaml = useMemo(() => generatePredeployYaml(config), [config]);
  const errors = useMemo(() => validateConfigBuilder(config), [config]);

  return (
    <section className="content-grid builder-grid">
      <Card title="Service Configuration">
        <ServiceForm config={config} onChange={setConfig} />
      </Card>
      <Card title="predeploy.yaml Preview">
        <YamlPreview errors={errors} yaml={yaml} />
      </Card>
    </section>
  );
}
