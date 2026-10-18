import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { Button, Input } from "../../../../../shared";
import styles from "./LoginForm.module.css";
import { useLogin } from "../../../api/queries";
import { ApiError } from "../../../../../shared/api/http/ApiError";

type FormState = {
  email: string;
  password: string;
  remember: boolean;
};

type Errors = Partial<Record<keyof FormState, string>>;

const MIN_PASSWORD_LENGTH = 12;



export function LoginForm() {
  const [form, setForm] = useState<FormState>({
    email: "",
    password: "",
    remember: false,
  });
  const [errors, setErrors] = useState<Errors>({});
  const [serverError, setServerError] = useState<string | null>(null);

  const navigate = useNavigate();
  const login = useLogin();

  const update = <K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) {
      setErrors((prev) => ({ ...prev, [key]: undefined }));
    }
  };

  const validate = (): Errors => {
    const next: Errors = {};

    if (!form.email.trim()) {
      next.email = "Введите email";
    } else if (!/^\S+@\S+\.\S+$/.test(form.email)) {
      next.email = "Некорректный email";
    }

    if (!form.password) {
      next.password = "Введите пароль";
    } else if (form.password.length < MIN_PASSWORD_LENGTH) {
      next.password = `Минимум ${MIN_PASSWORD_LENGTH} символов`;
    } 

    return next;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setServerError(null);

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    try {
      await login.mutateAsync({
        email: form.email.trim().toLowerCase(),
        password: form.password,
      });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.code) {
          case "INVALID_CREDENTIALS":
            setServerError("Неверный email или пароль");
            break;
          case "VALIDATION_ERROR":
            setServerError(
              "Проверьте email и пароль. Пароль — не менее 12 символов и не более 128 байт.",
            );
            break;
          case "CSRF_VALIDATION_FAILED":
            // Сюда логин попасть не должен, но на всякий случай
            setServerError(
              "Сессия устарела. Обновите страницу и попробуйте снова.",
            );
            break;
          case "REQUEST_TOO_LARGE":
            setServerError("Слишком большой запрос");
            break;
          case "INTERNAL_ERROR":
            setServerError("Что-то пошло не так. Попробуйте позже.");
            break;
          default:
            setServerError(err.message);
        }
      } else {
        setServerError("Не удалось войти. Попробуйте позже.");
      }
    }
  };

  const submitting = login.isPending;

  return (
    <form className={styles.form} onSubmit={handleSubmit} noValidate>
      {serverError && <div className={styles.error}>{serverError}</div>}

      <Input
        label="Email"
        type="email"
        autoComplete="email"
        placeholder="you@example.com"
        value={form.email}
        onChange={(e) => update("email", e.target.value)}
        error={errors.email}
      />

      <Input
        label="Пароль"
        type="password"
        autoComplete="current-password"
        placeholder="••••••••••••"
        value={form.password}
        onChange={(e) => update("password", e.target.value)}
        error={errors.password}
      />

      <div className={styles.row}>
        {/*
         TODO: "remember me"
          Login всегда ставит cookie с фиксированным Expires из result.ExpiresAt.
          
        */}
        <label className={styles.checkbox}>
          <input
            type="checkbox"
            checked={form.remember}
            onChange={(e) => update("remember", e.target.checked)}
          />
          <span>Запомнить меня</span>
        </label>

        <Link to="/forgot-password" className={styles.forgotLink}>
          Забыли пароль?
        </Link>
      </div>

      <Button
        type="submit"
        variant="primary"
        size="lg"
        disabled={submitting}
        className={styles.submit}
      >
        {submitting ? "Входим..." : "Войти"}
      </Button>
    </form>
  );
}
