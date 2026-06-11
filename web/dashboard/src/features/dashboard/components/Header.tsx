type HeaderProps = {
  onRefresh?: () => void;
};

export function Header({ onRefresh }: HeaderProps) {
  return (
    <header className="page-header">
      <div>
        <p className="eyebrow">Local validation control plane</p>
        <h1>PreDeploy Guard Dashboard</h1>
      </div>
      {onRefresh && (
        <button className="secondary-button" type="button" onClick={onRefresh}>
          Refresh
        </button>
      )}
    </header>
  );
}
