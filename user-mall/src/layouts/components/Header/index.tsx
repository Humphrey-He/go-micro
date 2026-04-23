import { useNavigate } from 'react-router-dom'
import { SearchOutlined, MessageOutlined } from '@ant-design/icons'
import { Badge } from 'antd'
import LanguageSwitcher from '@/components/LanguageSwitcher'

export function MobileHeader() {
  const navigate = useNavigate()

  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-white border-b border-gray-200">
      <div className="flex items-center justify-between px-4 h-12">
        {/* 搜索框 */}
        <div
          className="flex-1 flex items-center bg-gray-100 rounded-full px-3 py-1.5"
          onClick={() => navigate('/search')}
        >
          <SearchOutlined className="text-gray-400 mr-2" />
          <span className="text-gray-400 text-sm">搜索商品</span>
        </div>

        {/* 图标按钮 */}
        <div className="flex items-center gap-3 ml-4">
          <LanguageSwitcher />
          <span
            className="text-sm text-gray-500 cursor-pointer hover:text-purple-600 transition-colors"
            onClick={() => window.open('http://localhost:3000', '_blank')}
            title="管理后台"
          >
            后台
          </span>
          <MessageOutlined
            className="text-xl text-gray-600"
            onClick={() => navigate('/messages')}
          />
          <Badge count={0} size="small">
            <span
              className="text-xl text-gray-600 cursor-pointer"
              onClick={() => navigate('/cart')}
            >
              🛒
            </span>
          </Badge>
        </div>
      </div>
    </header>
  )
}