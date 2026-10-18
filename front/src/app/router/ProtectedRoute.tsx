import { Navigate, Outlet, useLocation } from "react-router";
import { useAuth } from "../../features/auth/hooks/useAuth";

export function ProtectedRoute() {
  const { isAuthenticated, isLoading } = useAuth();
  const location = useLocation();

  // Пока грузим /me — не решаем, гость или нет
  if (isLoading) {
    return <div>Loading…</div>; 
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return <Outlet />;
}