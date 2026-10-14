import { useMe } from "../api/queries";

export function useAuth() {
  const query = useMe();

  return {
    user: query.data?.user ?? null,
    isLoading: query.isLoading,
    isAuthenticated: !!query.data?.user,
    refetch: query.refetch,
  };
}