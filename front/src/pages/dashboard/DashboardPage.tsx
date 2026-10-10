import styles from './DashboardPage.module.css'

export function DashboardPage() {
  return (
    <>
      <main className={styles.main}>
        <div className={styles.inner}>
          <h1 className={styles.title}>Личный кабинет</h1>
          <p className={styles.subtitle}>
            Здесь будут ваши тесты, результаты и аналитика.
          </p>
        </div>
      </main>
    </>
  )
}