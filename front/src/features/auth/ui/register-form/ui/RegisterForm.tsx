import { useState, type FormEvent } from "react";
import { Button } from "../../../../../shared/index";
import { Input } from "../../../../../shared/index";
import styles from "./RegisterForm.module.css";
import { useNavigate } from "react-router";
import { useRegister, useLogin } from "../../../api/queries";
import { HttpError } from "../../../api/client";

type FormState = {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
  agree: boolean;
};

type Errors = Partial<Record<keyof FormState, string>>;

export function RegisterForm() {
const navigate = useNavigate();
  const register = useRegister();
  const login = useLogin();
  const [serverError, setServerError] = useState<string | null>(null);

  const [form, setForm] = useState<FormState>({
    name: "",
    email: "",
    password: "",
    confirmPassword: "",
    agree: false,
  });
  const [errors, setErrors] = useState<Errors>({});
  const [submitting, setSubmitting] = useState(false);

  const update = <K extends keyof FormState>(key: K, value: FormState[K]) => {
    setForm((prev) => ({ ...prev, [key]: value }));
    if (errors[key]) {
      setErrors((prev) => ({ ...prev, [key]: undefined }));
    }
  };

  const validate = (): Errors => {
    const next: Errors = {};

    if (!form.name.trim()) {
      next.name = "Введите имя";
    } else if (form.name.trim().length < 2) {
      next.name = "Минимум 2 символа";
    }

    if (!form.email.trim()) {
      next.email = "Введите email";
    } else if (!/^\S+@\S+\.\S+$/.test(form.email)) {
      next.email = "Некорректный email";
    }

    if (!form.password) {
      next.password = "Введите пароль";
    } else if (form.password.length < 8) {
      next.password = "Минимум 8 символов";
    } else if (!/[A-Za-z]/.test(form.password) || !/\d/.test(form.password)) {
      next.password = "Пароль должен содержать буквы и цифры";
    }

    if (!form.confirmPassword) {
      next.confirmPassword = "Подтвердите пароль";
    } else if (form.confirmPassword !== form.password) {
      next.confirmPassword = "Пароли не совпадают";
    }

    if (!form.agree) {
      next.agree = "Необходимо согласие с условиями";
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
      await register.mutateAsync({
        display_name: form.name.trim(),
        email: form.email.trim().toLowerCase(),
        password: form.password,
      });
      // После успешной регистрации, автоматически войти
      await login.mutateAsync({
        email: form.email.trim().toLowerCase(),
        password: form.password,
      });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof HttpError) {
        setServerError(err.message);
      } else {
        setServerError("Не удалось создать аккаунт. Попробуйте позже.");
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <form className={styles.form} onSubmit={handleSubmit} noValidate>
      {serverError && <div role="alert">{serverError}</div>}

      <Input
        label="Имя"
        type="text"
        autoComplete="name"
        placeholder="Как к вам обращаться"
        value={form.name}
        onChange={(e) => update("name", e.target.value)}
        error={errors.name}
      />

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
        autoComplete="new-password"
        placeholder="Минимум 8 символов"
        value={form.password}
        onChange={(e) => update("password", e.target.value)}
        error={errors.password}
        hint={
          !errors.password
            ? "Минимум 8 символов, содержит буквы и цифры"
            : undefined
        }
      />

      <Input
        label="Подтвердите пароль"
        type="password"
        autoComplete="new-password"
        placeholder="Повторите пароль"
        value={form.confirmPassword}
        onChange={(e) => update("confirmPassword", e.target.value)}
        error={errors.confirmPassword}
      />

      <div className={styles.terms}>
        <label className={styles.checkbox}>
          <input
            type="checkbox"
            checked={form.agree}
            onChange={(e) => update("agree", e.target.checked)}
          />
          <span>
            Я согласен с{" "}
            <a
              href="#"
              className={styles.termsLink}
              onClick={(e) => e.preventDefault()}
            >
              условиями использования
            </a>{" "}
            и{" "}
            <a
              href="#"
              className={styles.termsLink}
              onClick={(e) => e.preventDefault()}
            >
              политикой конфиденциальности
            </a>
          </span>
        </label>
        {errors.agree && (
          <span className={styles.termsError}>{errors.agree}</span>
        )}
      </div>

      <Button
        type="submit"
        variant="primary"
        size="lg"
        disabled={submitting}
        className={styles.submit}
      >
        {submitting ? "Создаём аккаунт..." : "Создать аккаунт"}
      </Button>
    </form>
  );
}
