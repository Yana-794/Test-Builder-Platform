import { Link } from 'react-router'
import notFoundImage from '../../assets/404_not_found.png'

export function NotFoundPage() {
  return (
    <main style={{ textAlign: 'center', padding: 32 }}>
      <img
        src={notFoundImage}
        alt="Страница не найдена"
        style={{
          maxWidth: '700px',
          width: '100%',
          height: 'auto',
          display: 'block',
          margin: '0 auto',
        }}
      />
      <h1>Страница не найдена</h1>
      <p>Возможно, ссылка устарела или в адресе опечатка</p>
      <Link to="/">Вернуться на главную</Link>
    </main>
  )
}