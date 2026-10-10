import styles from './TestPreview.module.css'

export function TestPreviewSection() {
  return (
    <section className={styles.section}>
      <div className={styles.inner}>
        <div className={styles.text}>
          <h2 className={styles.title}>
            Понятно вам. Удобно вашим ученикам.
          </h2>
          <p className={styles.subtitle}>
            Создавайте тесты в удобном редакторе, а ученики будут проходить их
            в простом интерфейсе без лишних шагов.
          </p>
        </div>

        <div className={styles.preview}>
          <div className={styles.previewCard}>
            <div className={styles.previewHeader}>
              <span className={styles.previewTitle}>Основы JavaScript</span>
              <span className={styles.previewCounter}>Вопрос 1 из 5</span>
            </div>

            <div className={styles.progress}>
              <div className={styles.progressFill} style={{ width: '20%' }} />
            </div>

            <p className={styles.question}>
              Как объявить переменную в JavaScript?
            </p>

            <div className={styles.options}>
              <div className={styles.option}>var</div>
              <div className={`${styles.option} ${styles.optionSelected}`}>
                <span className={styles.optionMarker} />
                const
              </div>
              <div className={styles.option}>let</div>
            </div>

            <div className={styles.previewFooter}>
              <button className={styles.previewButton} type="button">
                Далее →
              </button>
            </div>
          </div>

          <div className={styles.previewGlow} aria-hidden />
        </div>
      </div>
    </section>
  )
}