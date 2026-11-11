import type { FC } from "react";
import { useNavigate } from "react-router";
import { Plus } from "lucide-react";
import styles from "./DashboardPage.module.css";
import { ApiError } from "../../shared/api/http";
import { useAuth } from "../../features/auth/hooks/useAuth";
import { useDashboardOverview } from "../../features/dashboard-overview/model/useDashboardOverview";
import { StatsCards } from "../../features/dashboard-overview/ui/StatsCards/StatsCards";
import { RecentTests } from "../../features/dashboard-overview/ui/RecentTests/RecentTests";
import { ErrorState } from "../../shared/ui/error-state/ErrorState";

export const DashboardPage: FC = () => {
  const navigate = useNavigate();
  const { user } = useAuth();
  const userName = user?.display_name ?? "пользователь";

  const { data, isLoading, isError, error, refetch } = useDashboardOverview();

  const isUnauthorized = error instanceof ApiError && error.status === 401;
  const isNetwork = error instanceof ApiError && error.isNetworkError;

  const handleCreate = () => navigate("/app/tests/new");

  const handleCardClick = (filter: "all" | "draft" | "published") => {
    navigate(`/app/tests?filter=${filter}`);
  };

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <div className={styles.headerContent}>
          <h1 className={styles.title}>Добро пожаловать, {userName}!</h1>
          <p className={styles.subtitle}>
            Здесь вы можете управлять тестами и отслеживать результаты их
            прохождения.
          </p>
        </div>
        <button
          type="button"
          className={styles.primaryButton}
          onClick={handleCreate}
        >
          <Plus size={16} strokeWidth={2.5} />
          Создать тест
        </button>
      </header>

      {isLoading && (
        <>
          <StatsCards isLoading />
          <RecentTests isLoading onCreate={handleCreate} />
        </>
      )}

      {!isLoading && isUnauthorized && (
        <ErrorState
          message="Сессия истекла. Войдите заново."
          retryLabel="Войти"
          onRetry={() => navigate("/login")}
        />
      )}

      {!isLoading && !isUnauthorized && isError && (
        <ErrorState
          message={
            isNetwork
              ? "Нет соединения с сервером"
              : "Не удалось загрузить данные"
          }
          onRetry={() => refetch()}
        />
      )}

      {!isLoading && !isError && data && (
        <>
          <StatsCards
            isLoading={false}
            stats={data.stats}
            onCardClick={handleCardClick}
          />
          <RecentTests
            isLoading={false}
            items={data.recentTests}
            onCreate={handleCreate}
          />
        </>
      )}
    </div>
  );
};