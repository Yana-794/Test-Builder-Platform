import { useEffect, useState } from 'react'
import clsx from 'clsx'
import { Link } from 'react-router'
import { Button } from '../../../shared/index'
import { Logo } from '../../../shared/index'
import styles from './Header.module.css'

export function Header() {
  const [scrolled, setScrolled] = useState(false)
  const [menuOpen, setMenuOpen] = useState(false)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 8)
    window.addEventListener('scroll', onScroll, { passive: true })
    onScroll()
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  useEffect(() => {
    document.body.style.overflow = menuOpen ? 'hidden' : ''
    return () => {
      document.body.style.overflow = ''
    }
  }, [menuOpen])

  const handleAnchorClick = (
    e: React.MouseEvent<HTMLAnchorElement>,
    id: string
  ) => {
    e.preventDefault()
    const target = document.getElementById(id)
    if (target) {
      target.scrollIntoView({ behavior: 'smooth', block: 'start' })
      window.history.replaceState(null, '', `/#${id}`)
    }
    setMenuOpen(false)
  }

  return (
    <header className={clsx(styles.header, scrolled && styles.headerScrolled)}>
      <div className={styles.inner}>
        <Logo />
        <nav className={styles.nav}>
          <a
            href="/#features"
            className={styles.navLink}
            onClick={(e) => handleAnchorClick(e, 'features')}
          >
            Возможности
          </a>
          <a
            href="/#how-it-works"
            className={styles.navLink}
            onClick={(e) => handleAnchorClick(e, 'how-it-works')}
          >
            Как это работает
          </a>
        </nav>

        <div className={styles.actions}>
          <Button as="link" to="/login" variant="ghost" size="md">
            Войти
          </Button>
          <Button as="link" to="/register" variant="primary" size="md">
            Регистрация
          </Button>
        </div>

        <button
          className={styles.burger}
          aria-label={menuOpen ? 'Закрыть меню' : 'Открыть меню'}
          aria-expanded={menuOpen}
          onClick={() => setMenuOpen((v) => !v)}
        >
          <span className={clsx(styles.burgerLine, menuOpen && styles.burgerLineOpen1)} />
          <span className={clsx(styles.burgerLine, menuOpen && styles.burgerLineOpen2)} />
          <span className={clsx(styles.burgerLine, menuOpen && styles.burgerLineOpen3)} />
        </button>
      </div>

      {menuOpen && (
        <div className={styles.mobileMenu}>
          <nav className={styles.mobileNav}>
            <a
              href="/#features"
              className={styles.mobileLink}
              onClick={(e) => handleAnchorClick(e, 'features')}
            >
              Возможности
            </a>
            <a
              href="/#how-it-works"
              className={styles.mobileLink}
              onClick={(e) => handleAnchorClick(e, 'how-it-works')}
            >
              Как это работает
            </a>
            <Link to="/login" className={styles.mobileLink} onClick={() => setMenuOpen(false)}>
              Войти
            </Link>
            <Button as="link" to="/register" variant="primary" size="lg">
              Регистрация
            </Button>
          </nav>
        </div>
      )}
    </header>
  )
}