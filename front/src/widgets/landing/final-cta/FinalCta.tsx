import { Button } from '../../../shared/index'
import styles from './FinalCta.module.css'

export function FinalCtaSection() {
  // TODO: когда появится авторизация — вести на /app/tests/new, если вошёл
  const ctaLink = '/register'

  return (
    <section className={styles.section}>
      <div className={styles.inner}>
        <h2 className={styles.title}>Готовы создать свой первый тест?</h2>
        <p className={styles.subtitle}>
          Попробуйте собрать тест и поделиться им с учениками.
        </p>
        <div className={styles.actions}>
          <Button as="link" to={ctaLink} variant="secondary" size="lg">
            Начать создавать
          </Button>
        </div>
      </div>
    </section>
  )
}