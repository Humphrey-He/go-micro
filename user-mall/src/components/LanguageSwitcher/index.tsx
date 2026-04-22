import { useTranslation } from 'react-i18next'
import { Dropdown } from 'antd'
import type { MenuProps } from 'antd'

export default function LanguageSwitcher() {
  const { i18n } = useTranslation()

  const items: MenuProps['items'] = [
    {
      key: 'zh-CN',
      label: '中文',
      onClick: () => i18n.changeLanguage('zh-CN'),
    },
    {
      key: 'en-US',
      label: 'English',
      onClick: () => i18n.changeLanguage('en-US'),
    },
  ]

  const currentLang = i18n.language.startsWith('en') ? 'EN' : '中'

  return (
    <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
      <span className="cursor-pointer text-sm px-2 py-1 hover:bg-gray-100 rounded">
        {currentLang}
      </span>
    </Dropdown>
  )
}