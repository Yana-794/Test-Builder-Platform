import type { FC } from "react";
import { useState } from "react";
import styles from "./TestSettingsModal.module.css";
import type { EditorTestMeta } from "../../model/editorTypes";

type TestSettingsModalProps = {
  meta: EditorTestMeta;
  onChange: (patch: Partial<EditorTestMeta>) => void;
  onClose: () => void;
  disabled?: boolean;
};

export const TestSettingsModal: FC<TestSettingsModalProps> = ({
  meta,
  onChange,
  onClose,
  disabled,
}) => {
  const [local, setLocal] = useState<EditorTestMeta>(meta);

  const apply = () => {
    onChange({
      title: local.title.trim(),
      description: local.description.trim(),
      subject: local.subject?.trim() || null,
    });
    onClose();
  };

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <h2 className={styles.title}>Настройки теста</h2>

        <label className={styles.label}>
          Название *
          <input
            type="text"
            value={local.title}
            onChange={(e) => setLocal({ ...local, title: e.target.value })}
            maxLength={160}
            disabled={disabled}
          />
        </label>

        <label className={styles.label}>
          Описание
          <textarea
            value={local.description}
            onChange={(e) =>
              setLocal({ ...local, description: e.target.value })
            }
            maxLength={2000}
            rows={3}
            disabled={disabled}
          />
        </label>

        <label className={styles.label}>
          Предмет
          <input
            type="text"
            value={local.subject ?? ""}
            onChange={(e) =>
              setLocal({ ...local, subject: e.target.value || null })
            }
            maxLength={100}
            disabled={disabled}
          />
        </label>

        <div className={styles.actions}>
          <button className={styles.secondary} onClick={onClose}>
            Отмена
          </button>
          <button
            className={styles.primary}
            onClick={apply}
            disabled={disabled || !local.title.trim()}
          >
            Применить
          </button>
        </div>
      </div>
    </div>
  );
};
