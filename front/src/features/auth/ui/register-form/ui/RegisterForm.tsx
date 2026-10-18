import { useState, type FormEvent } from "react";
import { useNavigate } from "react-router";
import { Button, Input } from "../../../../../shared";
import styles from "./RegisterForm.module.css";
import { useLogin, useRegister } from "../../../api/queries";
import { ApiError } from "../../../../../shared/api/http/ApiError";
type FormState = {
  name: string;
  email: string;
  password: string;
  confirmPassword: string;
  agree: boolean;
};

type Errors = Partial<Record<keyof FormState, string>>;

const MIN_PASSWORD_LENGTH = 10;



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
    } else if (form.password.length < MIN_PASSWORD_LENGTH) {
      next.password = `Минимум ${MIN_PASSWORD_LENGTH} символов`;
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
    setServerError(null);

    const validationErrors = validate();
    if (Object.keys(validationErrors).length > 0) {
      setErrors(validationErrors);
      return;
    }

    const email = form.email.trim().toLowerCase();
    const displayName = form.name.trim();

    try {
      await register.mutateAsync({
        display_name: displayName,
        email,
        password: form.password,
      });
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.code) {
          case "EMAIL_ALREADY_REGISTERED":
            setServerError("Аккаунт с таким email уже существует");
            break;
          case "INTERNAL_ERROR":
            setServerError("Что-то пошло не так. Попробуйте позже.");
            break;
          default:
            setServerError(err.message);
        }
      } else {
        setServerError("Не удалось создать аккаунт. Попробуйте позже.");
      }
      return;
    }

    try {
      await login.mutateAsync({ email, password: form.password });
      navigate("/dashboard", { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        setServerError(
          "Аккаунт создан, но автоматический вход не удался. Войдите вручную.",
        );
      } else {
        setServerError("Аккаунт создан. Войдите, чтобы продолжить.");
      }
     
     navigate("/login", { replace: true });
    }
  };

  const submitting = register.isPending || login.isPending;

  return (
    <form className={styles.form} onSubmit={handleSubmit} noValidate>
      {serverError && (
        <div role="alert" className={styles.error}>
          {serverError}
        </div>
      )}

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
        placeholder={`Минимум ${MIN_PASSWORD_LENGTH} символов`}
        value={form.password}
        onChange={(e) => update("password", e.target.value)}
        error={errors.password}
        hint={
          !errors.password
            ? `От ${MIN_PASSWORD_LENGTH} символов. Используйте буквы, цифры и символы для надёжного пароля.`
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
