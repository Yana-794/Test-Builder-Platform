import { useState, type FC } from 'react'
import { Outlet, useNavigate } from 'react-router'

import { useAuth } from '../../features/auth/hooks/useAuth'
import styles from './AppLayout.module.css'
import { Header } from '../layout/header/Header'
import { Sidebar } from '../layout/sidebar/Sidebar'

export const AppLayout: FC = () => {
  const navigate = useNavigate()
  const { user, logout } = useAuth()
  const [isSidebarOpen, setSidebarOpen] = useState(false)

  const userName = user?.display_name ?? ''

  const handleLogout = () => {
    void logout()
  }

  return (
    <div className={styles.root}>
      <Header
        userName={userName}
        onCreateTest={() => navigate('/dashboard/tests/new')}
        onOpenSidebar={() => setSidebarOpen(true)}
        onLogout={handleLogout}
      />

      <Sidebar
        isOpen={isSidebarOpen}
        onClose={() => setSidebarOpen(false)}
      />

      <main className={styles.main}>
        <div className={styles.container}>
          <Outlet />
        </div>
      </main>
    </div>
  )
}