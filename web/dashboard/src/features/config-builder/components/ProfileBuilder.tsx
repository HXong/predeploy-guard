import type { FormEvent } from "react";
import {
  createCustomProfile,
  createLightLoadProfile,
  createSmokeOnlyProfile,
  createStressTestProfile,
} from "../defaults";
import type { ConfigBuilderState, ValidationProfileDraft } from "../types";
import { ProfileEditor } from "./ProfileEditor";

type ProfileBuilderProps = {
  config: ConfigBuilderState;
  onChange: (config: ConfigBuilderState) => void;
};

export function ProfileBuilder({ config, onChange }: ProfileBuilderProps) {
  function updateProfiles(profiles: ValidationProfileDraft[]) {
    onChange({
      ...config,
      profiles,
    });
  }

  function addProfile(profile: ValidationProfileDraft) {
    updateProfiles([...config.profiles, profile]);
  }

  function updateProfile(profile: ValidationProfileDraft) {
    updateProfiles(config.profiles.map((item) => (item.id === profile.id ? profile : item)));
  }

  function removeProfile(id: string) {
    updateProfiles(config.profiles.filter((profile) => profile.id !== id));
  }

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
  }

  return (
    <form className="builder-form" onSubmit={handleSubmit}>
      <p className="helper-text">
        Profile performance settings replace the complete base performance section when the profile is selected.
      </p>

      <div className="panel-actions">
        <button
          className="secondary-button"
          type="button"
          onClick={() => addProfile(createCustomProfile(config.performance))}
        >
          Add custom profile
        </button>
        <button className="secondary-button" type="button" onClick={() => addProfile(createSmokeOnlyProfile())}>
          Add smoke-only
        </button>
        <button className="secondary-button" type="button" onClick={() => addProfile(createLightLoadProfile())}>
          Add light-load
        </button>
        <button className="secondary-button" type="button" onClick={() => addProfile(createStressTestProfile())}>
          Add stress-test
        </button>
      </div>

      {config.profiles.length === 0 ? (
        <p className="empty-state">No validation profiles configured.</p>
      ) : (
        <div className="profile-list">
          {config.profiles.map((profile) => (
            <ProfileEditor
              key={profile.id}
              onChange={updateProfile}
              onRemove={() => removeProfile(profile.id)}
              profile={profile}
              smokeCheckCount={config.checks.smoke.length}
            />
          ))}
        </div>
      )}
    </form>
  );
}
