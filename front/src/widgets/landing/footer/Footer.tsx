import { Link } from "react-router";
import { Send } from "lucide-react";
import { Logo } from "../../../shared/index";
import styles from "./Footer.module.css";

const CONTACT_EMAIL = "hello@testbuilder.dev";

const SOCIALS = [{ icon: Send, href: "https://t.me", label: "Telegram" }];

export function Footer() {
  const handleAnchor = (e: React.MouseEvent<HTMLAnchorElement>, id: string) => {
    e.preventDefault();
    const target = document.getElementById(id);
    target?.scrollIntoView({ behavior: "smooth", block: "start" });
    window.history.replaceState(null, "", `/#${id}`);
  };

  return (
    <footer className={styles.footer}>
      <div className={styles.inner}>
        <div className={styles.top}>
          <div className={styles.brand}>
            <Logo />
            <p className={styles.brandText}>
              Простой способ создавать тесты и проверять знания.
            </p>

            <div className={styles.socials}>
              {SOCIALS.map((s) => {
                const Icon = s.icon;
                return (
                  <a
                    key={s.label}
                    href={s.href}
                    className={styles.socialBtn}
                    aria-label={s.label}
                    target="_blank"
                    rel="noopener noreferrer"
                  >
                    <Icon size={16} strokeWidth={2} />
                  </a>
                );
              })}
            </div>

            <div className={styles.status}>
              <span className={styles.statusDot} aria-hidden />
              <span>Все системы работают</span>
            </div>
          </div>

          <div className={styles.column}>
            <h4 className={styles.columnTitle}>Навигация</h4>
            <nav className={styles.links}>
              <a
                href="/#features"
                className={styles.link}
                onClick={(e) => handleAnchor(e, "features")}
              >
                Возможности
              </a>
              <a
                href="/#how-it-works"
                className={styles.link}
                onClick={(e) => handleAnchor(e, "how-it-works")}
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

          <div className={styles.column}>
            <h4 className={styles.columnTitle}>Контакты</h4>
            <div className={styles.links}>
              <a href={`mailto:${CONTACT_EMAIL}`} className={styles.link}>
                {CONTACT_EMAIL}
              </a>
            </div>
          </div>
        </div>

        <div className={styles.bottom}>
          <span className={styles.copyright}>
            © {new Date().getFullYear()} Test Builder Platform
          </span>
         
        </div>
      </div>
    </footer>
  );
}
