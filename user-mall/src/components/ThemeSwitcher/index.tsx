// user-mall/src/components/ThemeSwitcher/index.tsx
import { useLoginTheme } from '@/context/ThemeContext'
import { themes } from '@/styles/login-themes/theme-config'

export const ThemeSwitcher: React.FC = () => {
  const { theme, setTheme } = useLoginTheme()

  return (
    <div
      className="fixed top-4 right-4 z-50 flex items-center gap-3 px-4 py-2 rounded-full bg-white/90 backdrop-blur-sm shadow-lg"
      style={{ boxShadow: '0 4px 20px rgba(0,0,0,0.1)' }}
    >
      {Object.values(themes).map((t) => (
        <button
          key={t.id}
          onClick={() => setTheme(t.id)}
          className={`relative w-8 h-8 rounded-full transition-all duration-300 ${
            theme === t.id ? 'scale-110 ring-2 ring-offset-2' : 'opacity-60 hover:opacity-100'
          }`}
          style={{
            background: `linear-gradient(135deg, ${t.colors.primary}, ${t.colors.secondary})`,
          }}
          title={t.nameCn}
        >
          {theme === t.id && (
            <span className="absolute inset-0 flex items-center justify-center text-white text-xs font-bold">
              {t.nameCn.charAt(0)}
            </span>
          )}
        </button>
      ))}
    </div>
  )
}