import { useState, useCallback } from 'react'
import { Modal } from 'antd'

interface UseConfirmOptions {
  title?: string
  content?: string
  onConfirm?: () => Promise<void> | void
}

export const useConfirm = () => {
  const [loading, setLoading] = useState(false)

  const confirm = useCallback(
    (options: UseConfirmOptions) => {
      const { title = '确认操作', content = '确定要执行此操作吗？', onConfirm } = options

      Modal.confirm({
        title,
        content,
        okText: '确认',
        cancelText: '取消',
        okButtonProps: { danger: true, loading },
        onOk: async () => {
          if (!onConfirm) return
          setLoading(true)
          try {
            await onConfirm()
          } finally {
            setLoading(false)
          }
        },
      })
    },
    [loading]
  )

  return { confirm, loading }
}
