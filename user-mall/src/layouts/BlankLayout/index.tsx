import { Outlet } from 'react-router-dom'

export function BlankLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      {children}
      <Outlet />
    </div>
  )
}