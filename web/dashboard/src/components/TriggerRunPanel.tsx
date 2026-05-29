import { isSelectedProfile } from "../utils/format";
import { Card } from "./Card";

type TriggerRunPanelProps = {
  error: string | null;
  onProfileChange: (profile: string) => void;
  onTriggerRun: () => void;
  profile: string;
  triggering: boolean;
};

const profiles = ["base", "smoke-only", "light-load", "stress-test"];

export function TriggerRunPanel({
  error,
  onProfileChange,
  onTriggerRun,
  profile,
  triggering,
}: TriggerRunPanelProps) {
  return (
    <Card title="Trigger Run">
      <div className="profile-panel">
        <label htmlFor="profile">Profile (optional)</label>
        <input
          id="profile"
          value={profile}
          onChange={(event) => onProfileChange(event.target.value)}
          placeholder="Leave empty for base"
        />
        <p className="helper-text">Leave empty to run the base config.</p>
        <div className="profile-buttons">
          {profiles.map((candidate) => (
            <button
              className={isSelectedProfile(candidate, profile) ? "chip active" : "chip"}
              key={candidate}
              type="button"
              onClick={() => onProfileChange(candidate === "base" ? "" : candidate)}
            >
              {candidate}
            </button>
          ))}
        </div>
        <button className="primary-button" type="button" disabled={triggering} onClick={onTriggerRun}>
          {triggering ? "Starting..." : "Start validation"}
        </button>
        {error && <p className="error-text">{error}</p>}
      </div>
    </Card>
  );
}
