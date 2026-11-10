import { Fragment, useEffect, useRef } from "react";
import { PencilLine, Link2, BarChart3 } from "lucide-react";
import styles from "./HowItWorks.module.css";
import { SectionWordmark } from "../../../shared/ui/section-wordmark/SectionWordmark";

const STEPS = [
  {
    number: "01",
    icon: PencilLine,
    title: "Создайте тест",
    text: "Вопросы и баллы в простом редакторе",
  },
  {
    number: "02",
    icon: Link2,
    title: "Отправьте ссылку",
    text: "Ученики проходят без регистрации",
  },
  {
    number: "03",
    icon: BarChart3,
    title: "Изучите результат",
    text: "Баллы и статистика по каждому вопросу",
  },
];

export function HowItWorks() {
  const pathRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = pathRef.current;
    if (!el) return;

    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          el.classList.add(styles.pathVisible);
          observer.disconnect();
        }
      },
      { threshold: 0.15 },
    );

    observer.observe(el);
    return () => observer.disconnect();
  }, []);

  return (
    <section id="how-it-works" className={styles.section}>
       <SectionWordmark
              text="TESTBUILDER"
              lines={1}
              intensity={0.025}
              size="26vw"
            />
      <div className={styles.inner}>
        <div className={styles.head}>
          <div className={styles.headLeft}>
            <span className={styles.eyebrow}>
              <span className={styles.eyebrowDot} aria-hidden />
              Как это работает
            </span>
            <p className={styles.subtitle}>
              От создания теста до готовых результатов — весь путь в одном окне
            </p>
          </div>

          <h2 className={styles.title}>
            От идеи до результата{" "}
            <span className={styles.titleAccent}>за три шага</span>
          </h2>
        </div>

        <div ref={pathRef} className={styles.path}>
          {STEPS.map((step, i) => {
            const Icon = step.icon;
            return (
              <Fragment key={step.number}>
                <article className={styles.step}>
                  <span className={styles.stepNumber}>{step.number}</span>

                  <div className={styles.stepIcon}>
                    <Icon size={20} strokeWidth={2} />
                  </div>

                  <h3 className={styles.stepTitle}>{step.title}</h3>
                  <p className={styles.stepText}>{step.text}</p>
                </article>

                {i < STEPS.length - 1 && (
                  <div className={styles.connector} aria-hidden>
                    <span />
                    <span />
                    <span />
                  </div>
                )}
              </Fragment>
            );
          })}
        </div>
      </div>
    </section>
  );
}
