import { useState, type FormEvent } from "react";
import { Link, useNavigate } from "react-router";
import { Button } from "../../../../../shared/index";
import { Input } from "../../../../../shared/index";
import styles from "./LoginForm.module.css";
import { useLogin } from "../../../api/queries";
import { HttpError } from "../../../api/client";

type FormState = {
  email: string;
  password: string;
  remember: boolean;
};

type Errors = Partial<Record<keyof FormState, string>>;

export function LoginForm() {
  const [form, setForm] = useState<FormState>({
    email: "",
    password: "",
    remember: false,
  });
  const [errors, setErrors] = useState<Errors>({});
  const [submitting, setSubmitting] = useState(false);
  const [serverError, setServerError] = useState<string | null>(null);
  
  const update = <K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) {
      setErrors((prev) => ({ ...prev, [key]: undefined }));
    }
  };
    const navigate = useNavigate();

  const login = useLogin();

  const validate = (): Errors => {
    const next: Errors = {};

    if (!form.email.trim()) {
      next.email = "Введите email";
    } else if (!/^\S+@\S+\.\S+$/.test(form.email)) {
      next.email = "Некорректный email";
    }

    if (!form.password) {
      next.password = "Введите пароль";
    } else if (form.password.length < 8) {
      next.password = "Минимум 8 символов";
    }

    return next;
  };

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    setSubmitting(true);

     try {
      await login.mutateAsync({
        email: form.email.trim().toLowerCase(),
        password: form.password,
      });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof HttpError) {
        if (err.code === "INVALID_CREDENTIALS") {
          setServerError("Неверный email или пароль");
        } else if (err.code === "REQUEST_TOO_LARGE") {
          setServerError("Слишком большой запрос");
        } else {
          setServerError(err.message);
        }
      } else {
        setServerError("Не удалось войти. Попробуйте позже.");
      }
    } finally {
      setSubmitting(false);
    }
  };

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
        placeholder="••••••••"
        value={form.password}
        onChange={(e) => update("password", e.target.value)}
        error={errors.password}
      />

      <div className={styles.row}>
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
