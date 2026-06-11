import type { ReactNode } from "react";

type CardProps = {
  title: string;
  loading?: boolean;
  error?: string | null;
  children: ReactNode;
};

export function Card({ title, loading, error, children }: CardProps) {
  return (
    <article className="card">
      <div className="card-header">
        <h2>{title}</h2>
        {loading && <span className="loading-pill">Loading</span>}
      </div>
      {error ? <p className="error-text">{error}</p> : children}
    </article>
  );
}
