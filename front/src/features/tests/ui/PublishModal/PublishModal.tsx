import type { FC } from "react";
import { useEffect, useRef, useState } from "react";
import { QRCodeSVG } from "qrcode.react";
import { useValidateTest } from "../../model/hooks/useValidateTest";
import { usePublishTest } from "../../model/hooks/usePublishTest";
import type { Publication } from "../../../../entities/test/model/types";
import { ApiError } from "../../../../shared/api/http";
import styles from "./PublishModal.module.css";

type PublishModalProps = {
  testId: string;
  testTitle: string;
  testDescription: string;
  revision: number;
  onClose: () => void;
};

export const PublishModal: FC<PublishModalProps> = ({
  testId,
  testTitle,
  testDescription,
  revision,
  onClose,
}) => {
  const validate = useValidateTest(testId);
  const publish = usePublishTest(testId);

  const [published, setPublished] = useState<Publication | null>(null);
  const [copyLabel, setCopyLabel] = useState("Скопировать ссылку");
  const [publishError, setPublishError] = useState<string | null>(null);
  const qrRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    validate.mutate();
  }, []);

  const handlePublish = async () => {
    setPublishError(null);
    try {
      const resp = await publish.mutateAsync(revision);
      setPublished(resp.publication);
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.status) {
          case 409:
            setPublishError(
              "Черновик был изменён в другой вкладке. Закройте окно и обновите редактор.",
            );
            return;
          case 422:
            setPublishError(
              "Тест не готов к публикации. Проверьте список ошибок выше.",
            );
            return;
          case 404:
            setPublishError("Тест не найден.");
            return;
        }
      }
      setPublishError("Не удалось опубликовать. Попробуйте ещё раз.");
    }
  };

  const handleCopy = async () => {
    if (!published) return;
    try {
      await navigator.clipboard.writeText(published.publicUrl);
      setCopyLabel("Скопировано!");
      setTimeout(() => setCopyLabel("Скопировать ссылку"), 2000);
    } catch {
      setCopyLabel("Не удалось скопировать");
    }
  };

  const handleDownloadQr = () => {
    const svg = qrRef.current?.querySelector("svg");
    if (!svg) return;
    const xml = new XMLSerializer().serializeToString(svg);
    const blob = new Blob([xml], { type: "image/svg+xml;charset=utf-8" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `test-${published?.publicToken.slice(0, 8) ?? "qr"}.svg`;
    a.click();
    URL.revokeObjectURL(url);
  };

  const validation = validate.data;
  const isChecking = validate.isPending;
  const isPublishing = publish.isPending;
  const canPublish = !!validation?.valid && !isPublishing && !published;

  return (
    <div className={styles.overlay} onClick={onClose}>
      <div className={styles.modal} onClick={(e) => e.stopPropagation()}>
        <header className={styles.header}>
          <h2 className={styles.title}>
            {published ? "Тест опубликован" : "Публикация теста"}
          </h2>
          <button
            className={styles.closeButton}
            onClick={onClose}
            aria-label="Закрыть"
          >
            ×
          </button>
        </header>

        {!published && (
          <div className={styles.body}>
            <div className={styles.testInfo}>
              <h3 className={styles.testTitle}>
                {testTitle || "(без названия)"}
              </h3>
              {testDescription && (
                <p className={styles.testDesc}>{testDescription}</p>
              )}
            </div>

            {isChecking && (
              <div className={styles.loading}>
                <div className={styles.spinner} />
                Проверяем готовность теста...
              </div>
            )}

            {validate.isError && (
              <div className={styles.errorBox}>
                Не удалось проверить готовность теста. Попробуйте закрыть и
                открыть снова.
              </div>
            )}

            {validation && (
              <>
                <div className={styles.summary}>
                  <div className={styles.summaryItem}>
                    <span className={styles.summaryValue}>
                      {validation.summary.questionCount}
                    </span>
                    <span className={styles.summaryLabel}>вопросов</span>
                  </div>
                  <div className={styles.summaryItem}>
                    <span className={styles.summaryValue}>
                      {validation.summary.maxScore}
                    </span>
                    <span className={styles.summaryLabel}>макс. баллов</span>
                  </div>
                </div>

                {validation.valid ? (
                  <div className={styles.successBox}>
                    Тест готов к публикации
                  </div>
                ) : (
                  <div className={styles.errorsBox}>
                    <div className={styles.errorsTitle}>
                      Найдены ошибки ({validation.errors.length}):
                    </div>
                    <ul className={styles.errorsList}>
                      {validation.errors.map((err, i) => (
                        <li key={i} className={styles.errorItem}>
                          <span className={styles.errorField}>{err.field}</span>
                          <span className={styles.errorMessage}>
                            {err.message}
                          </span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}

                {publishError && (
                  <div className={styles.errorBox}>{publishError}</div>
                )}
              </>
            )}

            <div className={styles.actions}>
              <button
                className={styles.secondaryButton}
                onClick={onClose}
                disabled={isPublishing}
              >
                Отмена
              </button>
              <button
                className={styles.primaryButton}
                onClick={handlePublish}
                disabled={!canPublish}
              >
                {isPublishing ? "Публикация..." : "Опубликовать тест"}
              </button>
            </div>
          </div>
        )}

        {published && (
          <div className={styles.body}>
            <div className={styles.successBox}>
               Тест успешно опубликован (версия {published.version})
            </div>

            <div className={styles.qrWrap}>
              <div ref={qrRef} className={styles.qr}>
                <QRCodeSVG
                  value={published.publicUrl}
                  size={200}
                  includeMargin
                />
              </div>
              <button
                className={styles.downloadQrButton}
                onClick={handleDownloadQr}
              >
                Скачать QR
              </button>
            </div>

            <div className={styles.linkBox}>
              <input
                className={styles.linkInput}
                value={published.publicUrl}
                readOnly
                onFocus={(e) => e.currentTarget.select()}
              />
              <button className={styles.copyButton} onClick={handleCopy}>
                {copyLabel}
              </button>
            </div>

            <p className={styles.hint}>
              Отправьте эту ссылку или QR-код ученикам. Тест доступен без
              регистрации.
            </p>

            <div className={styles.actions}>
              <button className={styles.secondaryButton} onClick={onClose}>
                Закрыть
              </button>
              <button
                className={styles.primaryButton}
                onClick={() =>
                  window.open(
                    published.publicUrl,
                    "_blank",
                    "noopener,noreferrer",
                  )
                }
              >
                Открыть тест
              </button>
            </div>
          </div>
        )}
      </div>
    </div>
  );
};
