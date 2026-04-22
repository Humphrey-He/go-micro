import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { SearchBar } from 'antd-mobile'

interface HotSearch {
  id: string
  keyword: string
  heat: number
}

export default function Search() {
  const navigate = useNavigate()
  const [keyword, setKeyword] = useState('')
  const [suggestions, setSuggestions] = useState<string[]>([])
  const [hotSearches, setHotSearches] = useState<HotSearch[]>([])
  const [history, setHistory] = useState<string[]>([])

  useEffect(() => {
    loadHotSearches()
    loadHistory()
  }, [])

  useEffect(() => {
    if (keyword.length > 0) {
      loadSuggestions(keyword)
    } else {
      setSuggestions([])
    }
  }, [keyword])

  const loadHotSearches = async () => {
    setHotSearches([
      { id: '1', keyword: 'iPhone 15', heat: 9865 },
      { id: '2', keyword: '华为Mate60', heat: 8542 },
      { id: '3', keyword: '小米手机', heat: 7654 },
      { id: '4', keyword: 'AirPods Pro', heat: 6532 },
      { id: '5', keyword: 'MacBook', heat: 5432 },
      { id: '6', keyword: 'iPad', heat: 4321 },
    ])
  }

  const loadHistory = () => {
    const saved = localStorage.getItem('search_history')
    if (saved) {
      setHistory(JSON.parse(saved))
    }
  }

  const loadSuggestions = async (kw: string) => {
    const mock = [
      'iPhone 15 Pro',
      'iPhone 15 Pro Max',
      'iPhone 14',
      'iPhone 14 Pro',
      'iPhone 配件',
    ].filter((s) => s.toLowerCase().includes(kw.toLowerCase()))
    setSuggestions(mock)
  }

  const handleSearch = (kw: string) => {
    if (!kw.trim()) return
    const newHistory = [kw, ...history.filter((h) => h !== kw)].slice(0, 20)
    setHistory(newHistory)
    localStorage.setItem('search_history', JSON.stringify(newHistory))
    navigate(`/product/list?keyword=${encodeURIComponent(kw)}`)
  }

  const handleClearHistory = () => {
    setHistory([])
    localStorage.removeItem('search_history')
  }

  return (
    <div className="min-h-screen bg-gray-50">
      {/* 搜索栏 */}
      <div className="sticky top-0 z-50 bg-white p-2">
        <SearchBar
          placeholder="搜索商品"
          value={keyword}
          onChange={setKeyword}
          onSearch={handleSearch}
          onClear={() => setKeyword('')}
        />
      </div>

      {/* 搜索建议 */}
      {suggestions.length > 0 && (
        <div className="bg-white">
          {suggestions.map((s, i) => (
            <div
              key={i}
              className="px-4 py-3 border-b border-gray-100"
              onClick={() => handleSearch(s)}
            >
              {s}
            </div>
          ))}
        </div>
      )}

      {/* 搜索建议为空时显示默认内容 */}
      {suggestions.length === 0 && keyword === '' && (
        <>
          {/* 热搜榜单 */}
          <div className="bg-white mt-2 p-4">
            <div className="flex items-center gap-2 mb-3">
              <span className="text-xl">🔥</span>
              <span className="font-bold">热搜榜单</span>
            </div>
            <div className="flex flex-wrap gap-2">
              {hotSearches.map((item, index) => (
                <div
                  key={item.id}
                  className={`px-3 py-1.5 rounded-full text-sm ${
                    index < 3
                      ? 'bg-red-50 text-red-500'
                      : 'bg-gray-100 text-gray-600'
                  }`}
                  onClick={() => handleSearch(item.keyword)}
                >
                  {index < 3 ? '🔥' : ''} {item.keyword}
                </div>
              ))}
            </div>
          </div>

          {/* 搜索历史 */}
          {history.length > 0 && (
            <div className="bg-white mt-2 p-4">
              <div className="flex items-center justify-between mb-3">
                <div className="flex items-center gap-2">
                  <span className="text-xl">🕒</span>
                  <span className="font-bold">搜索历史</span>
                </div>
                <span
                  className="text-sm text-gray-400"
                  onClick={handleClearHistory}
                >
                  清空
                </span>
              </div>
              <div className="flex flex-wrap gap-2">
                {history.map((item, index) => (
                  <div
                    key={index}
                    className="px-3 py-1.5 rounded-full text-sm bg-gray-100 text-gray-600"
                    onClick={() => handleSearch(item)}
                  >
                    {item}
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* 常用分类快捷入口 */}
          <div className="bg-white mt-2 p-4">
            <div className="flex items-center gap-2 mb-3">
              <span className="text-xl">📞</span>
              <span className="font-bold">常用分类</span>
            </div>
            <div className="grid grid-cols-4 gap-4">
              {[
                { name: '手机', icon: '📱' },
                { name: '电脑', icon: '💻' },
                { name: '服装', icon: '👔' },
                { name: '美妆', icon: '💄' },
                { name: '家电', icon: '📺' },
                { name: '食品', icon: '🍎' },
                { name: '母婴', icon: '🍼' },
                { name: '家居', icon: '🏠' },
              ].map((cat) => (
                <div
                  key={cat.name}
                  className="flex flex-col items-center"
                  onClick={() => navigate(`/product/list?category=${cat.name}`)}
                >
                  <span className="text-2xl">{cat.icon}</span>
                  <span className="text-xs text-gray-600 mt-1">{cat.name}</span>
                </div>
              ))}
            </div>
          </div>
        </>
      )}
    </div>
  )
}