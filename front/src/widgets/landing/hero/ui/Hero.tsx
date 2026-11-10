import type { CSSProperties } from "react";
import { Link } from "react-router";
import mascotHero from "@/assets/mascot_hero.png";
import { useMouseParallax } from "../model/useMouseParallax";
import styles from "./Hero.module.css";
import { SectionWordmark } from "../../../../shared/ui/section-wordmark/SectionWordmark";

const TITLE_LINES = [
  { text: "Проверяй знания", from: "left" as const },
  { text: "без лишних шагов.", from: "right" as const },
];

const SUBTITLE = "Создавай тесты и отправляй по ссылке или QR. Ученики проходят без регистрации — проверка, оценки и аналитика готовы автоматически.";

const LETTER_STAGGER = 30;

interface HeroProps {
  mascotSrc?: string;
}

export function Hero({ mascotSrc = mascotHero }: HeroProps) {
  const heroRef = useMouseParallax();
  let letterIndex = 0;
  return (
    <section ref={heroRef} className={styles.hero}>
      <SectionWordmark
        text="TESTBUILDER"
        lines={5}
        intensity={0.025}
        size="26vw"
      />

      <div className={styles.blobs} aria-hidden>
        <span className={`${styles.blob} ${styles.blobA}`} />
        <span className={`${styles.blob} ${styles.blobB}`} />
        <span className={`${styles.blob} ${styles.blobC}`} />
        <span className={`${styles.blob} ${styles.blobD}`} />
      </div>

      <div className={styles.grid}>
        <div className={styles.text}>
          <h1
            className={styles.title}
            aria-label={TITLE_LINES.map((l) => l.text).join(" ")}
          >
            {TITLE_LINES.map((line, lineIdx) => (
              <span key={lineIdx} className={styles.line} aria-hidden="true">
                {Array.from(line.text).map((char, i) => {
                  const delay = letterIndex++ * LETTER_STAGGER;
                  const dirClass =
                    line.from === "left" ? styles.fromLeft : styles.fromRight;

                  return (
                    <span
                      key={i}
                      className={`${styles.letter} ${dirClass}`}
                      style={{ "--delay": `${delay}ms` } as CSSProperties}
                    >
                      {char === " " ? "\u00A0" : char}
                    </span>
                  );
                })}
              </span>
            ))}
          </h1>

          <div className={styles.answerOption}>
            <span className={styles.answerRadio} aria-hidden="true" />
            <p className={styles.answerText}>{SUBTITLE}</p>
          </div>

          <div className={styles.actions}>
            <Link to="/register" className={styles.btnPrimary}>
              <span>Создать тест</span>
              <svg
                className={styles.btnArrow}
                width="16"
                height="16"
                viewBox="0 0 24 24"
                fill="none"
                stroke="currentColor"
                strokeWidth="2.5"
                strokeLinecap="round"
                strokeLinejoin="round"
                aria-hidden="true"
              >
                <path d="M5 12h14M12 5l7 7-7 7" />
              </svg>
            </Link>

            <Link to="/login" className={styles.btnGhost}>
              Войти
            </Link>
          </div>

          <p className={styles.hint}>
            Бесплатно <span aria-hidden="true">·</span> Без карты
          </p>
        </div>

        <div className={styles.monsterStage}>
          <div className={styles.podium} aria-hidden>
            <div className={styles.podiumGlow} />
          </div>

          <img
            className={styles.monster}
            src={mascotSrc}
            alt=""
            draggable={false}
          />

          <aside className={styles.nameplate}>
            <span className={styles.nameplateArrow} aria-hidden />
            <span className={styles.nameplateTitle}>Это Пуша</span>
            <span className={styles.nameplateText}>проводник в мир тестов</span>
          </aside>
        </div>
      </div>
    </section>
  );
}
