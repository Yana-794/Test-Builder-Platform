import { useEffect, useState } from "react";
import { Link, useSearchParams } from "react-router";
import { AuthLoyaut } from "../../../widgets/auth-layout/AuthLoyaut";
import styles from "./VerifyEmailPage.module.css";

type Status = "loading" | "success" | "error";

export function VerifyEmailPage() {
  const [params] = useSearchParams();
  const token = params.get("token");
  const [status, setStatus] = useState<Status>(token ? "loading" : "error");

  useEffect(() => {
    if (!token) return;

    // TODO: заменить на реальный вызов API '/api/auth/verify-email'
    const timer = setTimeout(() => {
      setStatus(token === "invalid" ? "error" : "success");
    }, 800);

    return () => clearTimeout(timer);
  }, [token]);

  return (
    <AuthLoyaut
      title={
        status === "success"
          ? "Email подтверждён"
          : status === "error"
            ? "Не удалось подтвердить"
            : "Подтверждаем email"
      }
      subtitle={
        status === "success"
          ? "Теперь вы можете войти в аккаунт"
          : status === "error"
            ? "Ссылка недействительна или устарела"
            : "Это займёт пару секунд"
      }
      footer={
        status === "success" ? (
          <Link to="/login">Перейти ко входу</Link>
        ) : status === "error" ? (
          <Link to="/register">Зарегистрироваться </Link>
        ) : null
      }
    >
      <div className={styles.content}>
        {status === "loading" && <div className={styles.spinner} />}

        {status === "success" && (
          <div className={`${styles.icon} ${styles.iconSuccess}`}>✓</div>
        )}

        {status === "error" && (
          <div className={`${styles.icon} ${styles.iconError}`}>!</div>
        )}
      </div>
    </AuthLoyaut>
  );
}
