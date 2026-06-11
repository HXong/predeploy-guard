export type Resource<T> = {
  data: T | null;
  error: string | null;
  loading: boolean;
};

export function initialResource<T>(loading = true): Resource<T> {
  return {
    data: null,
    error: null,
    loading,
  };
}
