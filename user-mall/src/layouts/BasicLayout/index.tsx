import { MobileFooter } from '../components/Footer'
import { MobileHeader } from '../components/Header'

export function BasicLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen bg-gray-50">
      <MobileHeader />
      <main className="pb-16 pt-14">
        {children}
      </main>
      <MobileFooter />
    </div>
  )
}