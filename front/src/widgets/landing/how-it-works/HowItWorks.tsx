import { PencilLine, Link2, BarChart3 } from 'lucide-react'
import styles from './HowItWorks.module.css'

const STEPS = [
  {
    number: '01',
    icon: PencilLine,
    title: 'Создайте тест',
    text: 'Вопросы и баллы',
  },
  {
    number: '02',
    icon: Link2,
    title: 'Отправьте ссылку',
    text: 'Без аккаунта ученика',
  },
  {
    number: '03',
    icon: BarChart3,
    title: 'Изучите результат',
    text: 'Баллы и статистика',
  },
]

export function HowItWorks() {
  return (
    <section id="how-it-works" className={styles.section}>
      <div className={styles.inner}>
        <div className={styles.head}>
          <h2 className={styles.title}>
            От идеи до результата{' '}
            <span className={styles.titleAccent}>за три шага</span>
          </h2>
          <p className={styles.subtitle}>
            От создания теста до готовых результатов
          </p>
        </div>

        <div className={styles.path}>
          {STEPS.map((step) => {
            const Icon = step.icon
            return (
              <div key={step.number} className={styles.step}>
                <span className={styles.markerNumber}>{step.number}</span>

                <div className={styles.stepBody}>
                  <div className={styles.stepIcon}>
                    <Icon size={18} strokeWidth={2} />
                  </div>
                  <h3 className={styles.stepTitle}>{step.title}</h3>
                  <p className={styles.stepText}>{step.text}</p>
                </div>
              </div>
            )
          })}
        </div>
      </div>
    </section>
  )
}