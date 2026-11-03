import { useLogout, useMe } from "../api/queries";

export function useAuth() {
  const query = useMe();
  const logoutMutation = useLogout();
  return {
    user: query.data?.user ,
    isLoading: query.isLoading,
    isAuthenticated: !!query.data?.user,
    refetch: query.refetch,
    logout: logoutMutation.mutate,
    isLoggingOut: logoutMutation.isPending,
  };
}
