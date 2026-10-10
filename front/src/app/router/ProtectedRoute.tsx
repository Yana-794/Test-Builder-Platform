import type { ReactNode } from 'react'

type Props = {
  children: ReactNode
}

export function ProtectedRoute({ children }: Props) {
  // TODO: когда появится useAuth — проверить сессию
  // const { user, isLoading } = useAuth()
  // if (isLoading) return <FullPageSpinner />
  // if (!user) return <Navigate to="/login" replace state={{ from: location }} />

  return <>{children}</>
}