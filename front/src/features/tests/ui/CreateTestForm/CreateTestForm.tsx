import type { FC } from "react";
import { useState } from "react";
import { useNavigate } from "react-router";
import { useCreateTest } from "../../model/hooks/useCreateTest";
import { ApiError } from "../../../../shared/api/http";
import styles from "./CreateTestForm.module.css";

const TITLE_MAX = 160;
const DESCRIPTION_MAX = 2000;
const SUBJECT_MAX = 100;

export const CreateTestForm: FC = () => {
  const navigate = useNavigate();
  const mutation = useCreateTest();

  const [title, setTitle] = useState("");
  const [description, setDescription] = useState("");
  const [subject, setSubject] = useState("");
  const [fieldError, setFieldError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setFieldError(null);

    const normalizedTitle = title.trim();
    if (!normalizedTitle) {
      setFieldError("Название обязательно");
      return;
    }

    try {
      const test = await mutation.mutateAsync({
        title: normalizedTitle,
        description: description.trim(),
        subject: subject.trim() ? subject.trim() : null,
      });
      navigate(`/app/tests/${test.id}/edit`, { replace: true });
    } catch (err) {
      if (err instanceof ApiError) {
        if (err.status === 401) {
          navigate("/login", { replace: true });
          return;
        }
        if (err.status === 400) {
          setFieldError("Проверьте поля формы");
          return;
        }
      }
      setFieldError("Не удалось создать тест. Попробуйте ещё раз.");
    }
  };

  const isSubmitting = mutation.isPending;

  return (
    <form className={styles.form} onSubmit={handleSubmit} noValidate>
      <div className={styles.field}>
        <label className={styles.label} htmlFor="title">
          Название <span className={styles.required}>*</span>
        </label>
        <input
          id="title"
          type="text"
          className={styles.input}
          value={title}
          maxLength={TITLE_MAX}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Например, Основы JavaScript"
          disabled={isSubmitting}
          autoFocus
        />
        <span className={styles.counter}>
          {title.length} / {TITLE_MAX}
        </span>
      </div>

      <div className={styles.field}>
        <label className={styles.label} htmlFor="description">
          Описание
        </label>
        <textarea
          id="description"
          className={styles.textarea}
          value={description}
          maxLength={DESCRIPTION_MAX}
          onChange={(e) => setDescription(e.target.value)}
          placeholder="Добавьте описание теста..."
          rows={4}
          disabled={isSubmitting}
        />
        <span className={styles.counter}>
          {description.length} / {DESCRIPTION_MAX}
        </span>
      </div>

      <div className={styles.field}>
        <label className={styles.label} htmlFor="subject">
          Предмет
        </label>
        <input
          id="subject"
          type="text"
          className={styles.input}
          value={subject}
          maxLength={SUBJECT_MAX}
          onChange={(e) => setSubject(e.target.value)}
          placeholder="Например, Английский язык"
          disabled={isSubmitting}
        />
      </div>

      {fieldError && <div className={styles.error}>{fieldError}</div>}

      <div className={styles.actions}>
        <button
          type="button"
          className={styles.cancelButton}
          onClick={() => navigate("/app/tests")}
          disabled={isSubmitting}
        >
          Отмена
        </button>
        <button
          type="submit"
          className={styles.submitButton}
          disabled={isSubmitting}
        >
          {isSubmitting ? "Создание..." : "Создать тест"}
        </button>
      </div>
    </form>
  );
};
