import { useEffect, useRef } from 'react'
import { Button } from '../../../shared/index'
import styles from './FinalCta.module.css'
import { SectionWordmark } from '../../../shared/ui/section-wordmark/SectionWordmark'

export function FinalCtaSection() {
  const innerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = innerRef.current
    if (!el) return

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.classList.add(styles.innerVisible)
          observer.disconnect()
        }
      },
      { threshold: 0.3 }
    )

    observer.observe(el)
    return () => observer.disconnect()
  }, [])

  const ctaLink = '/register'

  return (
    <section className={styles.section}>
       <SectionWordmark
                    text="READYREADYREADY"
                    lines={5}
                    intensity={0.025}
                    size="26vw"
                    fade={false}
                  />

      <div className={styles.deco} aria-hidden>
        <span className={styles.glowTop} />
        <span className={styles.glowBottom} />
        <span className={styles.glowCenter} />
      </div>

      <div ref={innerRef} className={styles.inner}>
        <h2 className={styles.title}>Готовы создать свой первый тест?</h2>
        <p className={styles.subtitle}>
          Попробуйте собрать тест и поделиться им с учениками.
        </p>
        <div className={styles.actions}>
          <Button
            as="link"
            to={ctaLink}
            variant="secondary"
            size="lg"
            className={styles.ctaButton}
          >
            Начать создавать
          </Button>
        </div>
      </div>
    </section>
  )
}