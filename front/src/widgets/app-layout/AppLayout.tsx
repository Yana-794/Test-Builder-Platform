import type { FC } from "react";
import { Outlet, useMatch, useNavigate } from "react-router";
import { useAuth } from "../../features/auth/hooks/useAuth";
import styles from "./AppLayout.module.css";
import { Header } from "../layout/header/Header";
import { Sidebar } from "../layout/sidebar/Sidebar";

export const AppLayout: FC = () => {
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const isEditor = useMatch("/app/tests/:testId/edit") !== null;
  const userName = user?.display_name ?? "";

  const handleLogout = () => {
    void logout();
  };

  return (
    <div className={styles.root}>
      <Header
        userName={userName}
        onCreateTest={() => navigate("/app/tests/new")}
        onLogout={handleLogout}
      />

      <Sidebar />

      <main className={styles.main}>
        <div className={isEditor ? styles.containerFull : styles.container}>
          <Outlet />
        </div>
      </main>
    </div>
  );
};