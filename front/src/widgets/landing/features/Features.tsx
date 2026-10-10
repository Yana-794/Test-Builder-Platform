import { useEffect, useRef } from 'react'
import { CheckCircle2, Link2, BarChart3, PencilLine } from 'lucide-react'
import clsx from 'clsx'
import styles from './Features.module.css'

const FEATURES = [
  {
    icon: Link2,
    title: 'Делитесь  ссылкой',
    text: 'Отправляйте тест ученикам или показывайте QR-код.',
  },
  {
    icon: BarChart3,
    title: 'Смотрите результаты',
    text: 'Изучайте баллы, процент правильных ответов и статистику по вопросам.',
  },
]

export function Features() {
  const gridRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = gridRef.current
    if (!el) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.classList.add(styles.gridVisible)
          observer.disconnect()
        }
      },
      { threshold: 0.15 }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  return (
    <section id="features" className={styles.section}>
      <div className={styles.inner}>
        <div className={styles.head}>
          <h2 className={styles.title}>Всё необходимое для проверки знаний</h2>
          <p className={styles.subtitle}>
            Простой инструмент, который закрывает весь цикл — от создания теста
            до аналитики результатов.
          </p>
        </div>

        <div ref={gridRef} className={styles.grid}>
          {/* === Большая карточка === */}
          <article className={styles.cardMain}>
            <div className={styles.cardMainTextWrap}>
              <div className={styles.iconWrap}>
                <PencilLine size={22} strokeWidth={2} />
              </div>
              <h3 className={styles.cardMainTitle}>Создавайте тесты под свои задачи</h3>
              <p className={styles.cardMainDescription}>
              Добавляйте вопросы, варианты ответов и баллы в простом редакторе.
              </p>
            </div>

            <div className={styles.editorPreview} aria-hidden>
              <div className={styles.editorHeader}>
                <span className={styles.editorDot} />
                <span className={styles.editorDot} />
                <span className={styles.editorDot} />
              </div>

              <div className={styles.editorBody}>
                <div className={styles.editorQuestion}>
                  Сколько будет 8 × 7?
                </div>

                <div className={styles.editorOptions}>
                  <div className={styles.editorOption}>48</div>
                  <div
                    className={clsx(
                      styles.editorOption,
                      styles.editorOptionCorrect
                    )}
                  >
                    <CheckCircle2 size={14} strokeWidth={2.5} />
                    56
                  </div>
                  <div className={styles.editorOption}>54</div>
                </div>

                <div className={styles.editorMeta}>
                  <span className={styles.editorPoints}>1 балл</span>
                  <span className={styles.editorType}>
                    Один правильный ответ
                  </span>
                </div>
              </div>
            </div>
          </article>

          {/* === Маленькие карточки === */}
          {FEATURES.map((feature) => {
            const Icon = feature.icon
            return (
              <article key={feature.title} className={styles.card}>
                <div className={styles.iconWrap}>
                  <Icon size={22} strokeWidth={2} />
                </div>
                <h3 className={styles.cardTitle}>{feature.title}</h3>
                <p className={styles.cardText}>{feature.text}</p>
              </article>
            )
          })}
        </div>
      </div>
    </section>
  )
}