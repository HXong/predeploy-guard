import type { ValidationProfileDraft } from "../types";
import { PerformanceConfigEditor } from "./PerformanceConfigEditor";

type ProfileEditorProps = {
  profile: ValidationProfileDraft;
  smokeCheckCount: number;
  onChange: (profile: ValidationProfileDraft) => void;
  onRemove: () => void;
};

export function ProfileEditor({ profile, smokeCheckCount, onChange, onRemove }: ProfileEditorProps) {
  return (
    <section className="profile-card">
      <div className="section-header">
        <h3>{profile.name || "Validation profile"}</h3>
        <button className="secondary-button" type="button" onClick={onRemove}>
          Remove
        </button>
      </div>

      <label className="profile-name-field">
        Profile name
        <input value={profile.name} onChange={(event) => onChange({ ...profile, name: event.target.value })} />
      </label>

      <PerformanceConfigEditor
        onChange={(performance) => onChange({ ...profile, performance })}
        performance={profile.performance}
        smokeCheckCount={smokeCheckCount}
      />
    </section>
  );
}
