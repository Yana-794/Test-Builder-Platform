import { Link } from 'react-router'
import { Logo } from '../../../shared/index'
import styles from './Footer.module.css'

const CONTACT_EMAIL = 'hello@testbuilder.dev'

export function Footer() {
  const handleAnchor = (
    e: React.MouseEvent<HTMLAnchorElement>,
    id: string
  ) => {
    e.preventDefault()
    const target = document.getElementById(id)
    target?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    window.history.replaceState(null, '', `/#${id}`)
  }

  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <div className={styles.top}>
          {/* Колонка 1: бренд */}
          <div className={styles.brand}>
            <Logo />
            <p className={styles.brandText}>
              Простой способ создавать тесты и проверять знания.
            </p>
          </div>

          {/* Колонка 2: навигация */}
          <div className={styles.column}>
            <h4 className={styles.columnTitle}>Навигация</h4>
            <nav className={styles.links}>
              <a
                href="/#features"
                className={styles.link}
                onClick={(e) => handleAnchor(e, 'features')}
              >
                Возможности
              </a>
              <a
                href="/#how-it-works"
                className={styles.link}
                onClick={(e) => handleAnchor(e, 'how-it-works')}
              >
                Как это работает
              </a>
              <Link to="/login" className={styles.link}>
                Войти
              </Link>
              <Link to="/register" className={styles.link}>
                Регистрация
              </Link>
            </nav>
          </div>

          {/* Колонка 3: контакты */}
          <div className={styles.column}>
            <h4 className={styles.columnTitle}>Контакты</h4>
            <div className={styles.links}>
              <a href={`mailto:${CONTACT_EMAIL}`} className={styles.link}>
                {CONTACT_EMAIL}
              </a>
              <a href="#" className={styles.link}>
                Обратная связь
              </a>
            </div>
          </div>
        </div>

        <div className={styles.bottom}>
          <span className={styles.copyright}>
            © {new Date().getFullYear()} Test Builder Platform
          </span>
          <div className={styles.legal}>
            <a href="#" className={styles.legalLink}>
              Политика конфиденциальности
            </a>
            <span className={styles.separator} aria-hidden>
              ·
            </span>
            <a href="#" className={styles.legalLink}>
              Условия использования
            </a>
          </div>
        </div>
      </div>
    </footer>
  )
}