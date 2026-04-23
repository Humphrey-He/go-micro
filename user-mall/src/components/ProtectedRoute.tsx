import { Navigate } from 'react-router-dom'
import { useAuthStore } from '@/stores/authStore'

interface Props {
  children: React.ReactNode
  redirectTo?: string
}

export default function ProtectedRoute({ children, redirectTo = '/login' }: Props) {
  const { isLoggedIn } = useAuthStore()

  if (!isLoggedIn) {
    return <Navigate to={redirectTo} replace />
  }

  return <>{children}</>
}
