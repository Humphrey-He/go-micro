import { useNavigate, useLocation } from 'react-router-dom'

const tabs = [
  { key: '/', icon: '🏠', label: '首页' },
  { key: '/product/list', icon: '📱', label: '分类' },
  { key: '/seckill', icon: '⚡', label: '秒杀' },
  { key: '/cart', icon: '🛒', label: '购物车' },
  { key: '/user', icon: '👤', label: '我的' },
]

export function MobileFooter() {
  const navigate = useNavigate()
  const location = useLocation()

  return (
    <footer className="fixed bottom-0 left-0 right-0 z-50 bg-white border-t border-gray-200 safe-area-bottom">
      <div className="flex justify-around items-center h-14">
        {tabs.map((tab) => (
          <div
            key={tab.key}
            className={`flex flex-col items-center justify-center w-16 h-full cursor-pointer ${
              location.pathname === tab.key ? 'text-primary-500' : 'text-gray-400'
            }`}
            onClick={() => navigate(tab.key)}
          >
            <span className="text-xl">{tab.icon}</span>
            <span className="text-xs mt-0.5">{tab.label}</span>
          </div>
        ))}
      </div>
    </footer>
  )
}