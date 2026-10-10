import { Button } from '../../../shared//index'
// import lending from '../../../asset/lending.png'
import styles from './Hero.module.css'
import video from '../../../asset/video.webm'
export function Hero() {
  const handleHowItWorks = (e: React.MouseEvent<HTMLAnchorElement>) => {
    e.preventDefault()
    const target = document.getElementById('how-it-works')
    target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    window.history.replaceState(null, '', '/#how-it-works')
  }

  return (
    <section className={styles.hero}>
      <div className={styles.text}>
        <div className={styles.textInner}>
          <span className={styles.badge}>Тесты без лишней сложности</span>

          <h1 className={styles.title}>Создавайте тесты за минуты</h1>

          <p className={styles.subtitle}>
            Создавайте тесты для занятий, делитесь ими по ссылке и автоматически
            проверяйте ответы. Всё необходимое для проверки знаний — в одном месте.
          </p>

          <div className={styles.actions}>
            <Button as="link" to="/register" variant="primary" size="lg">
              Создать первый тест
            </Button>
            <a
              href="/#how-it-works"
              className={styles.secondaryLink}
              onClick={handleHowItWorks}
            >
              Как это работает ↓
            </a>
          </div>
        </div>
      </div>

      <div className={styles.visual}>
    <video
          className={styles.video}
          src={video}
          autoPlay
          loop
          muted
          playsInline
          preload="metadata"
          aria-hidden
        />
      </div>
    </section>
  )
}