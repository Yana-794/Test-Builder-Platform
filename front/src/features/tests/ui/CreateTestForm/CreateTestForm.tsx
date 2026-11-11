import type { FC } from "react";
import { useState } from "react";
import { useNavigate } from "react-router";
import {
  AlertCircle,
  AlignLeft,
  BookOpen,
  Hash,
  Loader2,
  Plus,
  X,
} from "lucide-react";
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

  const titleNear = title.length >= TITLE_MAX * 0.9;
  const descNear = description.length >= DESCRIPTION_MAX * 0.9;

  return (
    <form className={styles.form} onSubmit={handleSubmit} noValidate>
      <div className={styles.formBody}>
        <div className={styles.field}>
          <label className={styles.label} htmlFor="title">
            <span className={styles.labelText}>
              <Hash size={13} strokeWidth={2.4} aria-hidden="true" />
              Название
              <span className={styles.required} aria-hidden="true">
                *
              </span>
            </span>
            <span className={styles.labelHint}>Обязательно для создания</span>
          </label>
          <input
            id="title"
            type="text"
            className={styles.input}
            value={title}
            maxLength={TITLE_MAX}
            onChange={(e) => setTitle(e.target.value)}
            placeholder="Например, Основы Go"
            disabled={isSubmitting}
            autoFocus
            aria-required="true"
            aria-invalid={!!fieldError}
            aria-describedby="title-counter"
          />
          <span
            id="title-counter"
            className={`${styles.counter} ${titleNear ? styles.counterNear : ""}`}
          >
            {title.length} / {TITLE_MAX}
          </span>
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="description">
            <span className={styles.labelText}>
              <AlignLeft size={13} strokeWidth={2.4} aria-hidden="true" />
              Описание
            </span>
            <span className={styles.labelHint}>Необязательно</span>
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
            aria-describedby="description-counter"
          />
          <span
            id="description-counter"
            className={`${styles.counter} ${descNear ? styles.counterNear : ""}`}
          >
            {description.length} / {DESCRIPTION_MAX}
          </span>
        </div>

        <div className={styles.field}>
          <label className={styles.label} htmlFor="subject">
            <span className={styles.labelText}>
              <BookOpen size={13} strokeWidth={2.4} aria-hidden="true" />
              Предмет
            </span>
            <span className={styles.labelHint}>Необязательно</span>
          </label>
          <input
            id="subject"
            type="text"
            className={styles.input}
            value={subject}
            maxLength={SUBJECT_MAX}
            onChange={(e) => setSubject(e.target.value)}
            placeholder="Например, Программирование"
            disabled={isSubmitting}
          />
        </div>

        {fieldError && (
          <div className={styles.error} role="alert" aria-live="assertive">
            <AlertCircle size={18} strokeWidth={2.2} aria-hidden="true" />
            <span>{fieldError}</span>
          </div>
        )}
      </div>

      <div className={styles.actions}>
        <button
          type="button"
          className={styles.cancelButton}
          onClick={() => navigate("/app/tests")}
          disabled={isSubmitting}
        >
          <X size={15} strokeWidth={2.2} />
          Отмена
        </button>
        <button
          type="submit"
          className={styles.submitButton}
          disabled={isSubmitting || !title.trim()}
        >
          {isSubmitting ? (
            <Loader2 size={15} strokeWidth={2.4} className={styles.spin} />
          ) : (
            <Plus size={15} strokeWidth={2.4} />
          )}
          <span>{isSubmitting ? "Создание..." : "Создать тест"}</span>
        </button>
      </div>
    </form>
  );
};