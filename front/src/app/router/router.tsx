import { createBrowserRouter } from 'react-router'
import { LandingPage } from '../../pages/landing/LandingPage'
import { LoginPage } from '../../pages/auth/login/LoginPage'
import { RegisterPage } from '../../pages/auth/register/RegisterPage'
import { VerifyEmailPage } from '../../pages/auth/verify-email/VerifyEmailPage'
import { NotFoundPage } from '../../pages/not-found/NotFoundPage'
import { DashboardPage } from '../../pages/dashboard/DashboardPage'
import { ProtectedRoute } from './ProtectedRoute'

export const router = createBrowserRouter([
  { path: '/', element: <LandingPage /> },
  { path: '/login', element: <LoginPage /> },
  { path: '/register', element: <RegisterPage /> },
  { path: '/verify-email', element: <VerifyEmailPage /> },
  {
    path: '/dashboard',
    element: <ProtectedRoute />,
    children: [{ index: true, element: <DashboardPage /> }],
  },
  { path: '*', element: <NotFoundPage /> },
])