import { Link } from 'react-router'

export function NotFoundPage() {
  return (
    <main style={{ textAlign: 'center', padding: 32 }}>
      <div style={{ fontSize: 64, fontWeight: 700, color: 'var(--color-primary)', lineHeight: 1 }}>
        404
      </div>
      <h1>Страница не найдена</h1>
      <p>Возможно, ссылка устарела или в адресе опечатка</p>
      <Link to="/">Вернуться на главную</Link>
    </main>
  )
}