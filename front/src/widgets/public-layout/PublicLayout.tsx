import type { FC } from "react";
import { Outlet } from "react-router";
import styles from "./PublicLayout.module.css";

export const PublicLayout: FC = () => (
  <div className={styles.root}>
    <main className={styles.main}>
      <Outlet />
    </main>
    <footer className={styles.footer}>Test Builder Platform</footer>
  </div>
);