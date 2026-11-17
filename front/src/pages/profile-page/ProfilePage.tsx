import type { FC } from "react";
import { useState } from "react";
import { useAuth } from "../../features/auth/hooks/useAuth";
import { useUpdateProfile, useChangePassword } from "../../features/auth/api/queries";
import { ApiError } from "../../shared/api/http";
import type { User } from "../../features/auth/api/types";
import styles from "./ProfilePage.module.css";

export const ProfilePage: FC = () => {
  const { user, isLoading } = useAuth();

  if (isLoading || !user) {
    return <div className={styles.loading}>Загрузка...</div>;
  }

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <h1 className={styles.title}>Профиль</h1>
        <p className={styles.subtitle}>Управление аккаунтом и паролем</p>
      </header>

      <div className={styles.grid}>
        <ProfileForm user={user} />
        <ChangePasswordForm />
      </div>
    </div>
  );
};


const ProfileForm: FC<{ user: User }> = ({ user }) => {
  const [name, setName] = useState(user.display_name);
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const mutation = useUpdateProfile();

  const isDirty = name.trim() !== user.display_name;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSaved(false);
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Имя не может быть пустым");
      return;
    }
    try {
      await mutation.mutateAsync({ display_name: trimmed });
      setSaved(true);
      setTimeout(() => setSaved(false), 2000);
    } catch (err) {
      if (err instanceof ApiError && err.status === 400) {
        setError("Проверьте имя (1–100 символов)");
      } else {
        setError("Не удалось сохранить. Попробуйте ещё раз.");
      }
    }
  };

  return (
    <form className={styles.card} onSubmit={handleSubmit}>
      <h2 className={styles.cardTitle}>Аккаунт</h2>

      <label className={styles.label}>
        Email
        <input className={styles.input} value={user.email} readOnly disabled />
        <span className={styles.hint}>Используется для входа, изменить нельзя</span>
      </label>

      <label className={styles.label}>
        Отображаемое имя
        <input
          className={styles.input}
          value={name}
          maxLength={100}
          onChange={(e) => { setName(e.target.value); setError(null); setSaved(false); }}
          disabled={mutation.isPending}
        />
      </label>

      <p className={styles.meta}>
        Аккаунт создан {new Date(user.created_at).toLocaleDateString("ru-RU")}
      </p>

      {error && <div className={styles.error}>{error}</div>}
      {saved && <div className={styles.success}>Сохранено</div>}

      <div className={styles.actions}>
        <button type="submit" className={styles.primary} disabled={!isDirty || mutation.isPending}>
          {mutation.isPending ? "Сохранение..." : "Сохранить"}
        </button>
      </div>
    </form>
  );
};


const ChangePasswordForm: FC = () => {
  const [current, setCurrent] = useState("");
  const [next, setNext] = useState("");
  const [confirm, setConfirm] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [saved, setSaved] = useState(false);
  const mutation = useChangePassword();

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSaved(false);

    if (next.length < 12) { setError("Новый пароль — минимум 12 символов"); return; }
    if (next !== confirm) { setError("Пароли не совпадают"); return; }

    try {
      await mutation.mutateAsync({ current_password: current, new_password: next });
      setSaved(true);
      setCurrent(""); setNext(""); setConfirm("");
      setTimeout(() => setSaved(false), 3000);
    } catch (err) {
      if (err instanceof ApiError && err.status === 400) {
        setError("Текущий пароль неверный или новый слишком короткий");
      } else {
        setError("Не удалось сменить пароль");
      }
    }
  };

  return (
    <form className={styles.card} onSubmit={handleSubmit}>
      <h2 className={styles.cardTitle}>Смена пароля</h2>

      <label className={styles.label}>
        Текущий пароль
        <input type="password" className={styles.input} value={current}
          autoComplete="current-password"
          onChange={(e) => setCurrent(e.target.value)} />
      </label>
      <label className={styles.label}>
        Новый пароль (мин. 12)
        <input type="password" className={styles.input} value={next}
          autoComplete="new-password"
          onChange={(e) => setNext(e.target.value)} />
      </label>
      <label className={styles.label}>
        Повторите новый пароль
        <input type="password" className={styles.input} value={confirm}
          autoComplete="new-password"
          onChange={(e) => setConfirm(e.target.value)} />
      </label>

      {error && <div className={styles.error}>{error}</div>}
      {saved && <div className={styles.success}>Пароль изменён</div>}

      <p className={styles.hint}>
        После смены пароля все другие сессии будут завершены. Текущая останется активной.
      </p>

      <div className={styles.actions}>
        <button type="submit" className={styles.primary}
          disabled={mutation.isPending || !current || !next || !confirm}>
          {mutation.isPending ? "Смена..." : "Сменить пароль"}
        </button>
      </div>
    </form>
  );
};