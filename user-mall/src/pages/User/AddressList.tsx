import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Button, Empty, SwipeAction, Skeleton } from 'antd-mobile'
import { DeleteOutline, EditSOutline } from 'antd-mobile-icons'
import type { Address as AddressType } from '@/api/address'

const mockAddresses: AddressType[] = [
  {
    id: '1',
    receiver: '张三',
    phone: '138****8888',
    province: '广东省',
    city: '深圳市',
    district: '南山区',
    detail: '科技园南路88号',
    tag: 'home',
    is_default: true,
  },
  {
    id: '2',
    receiver: '李四',
    phone: '139****9999',
    province: '广东省',
    city: '广州市',
    district: '天河区',
    detail: '天河路123号',
    tag: 'company',
    is_default: false,
  },
]

export default function AddressList() {
  const navigate = useNavigate()
  const [addresses, setAddresses] = useState<AddressType[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    loadAddresses()
  }, [])

  const loadAddresses = async () => {
    try {
      setLoading(true)
      // 模拟数据
      setAddresses(mockAddresses)
    } finally {
      setLoading(false)
    }
  }

  const handleDelete = (id: string) => {
    setAddresses((prev) => prev.filter((a) => a.id !== id))
  }

  const handleSetDefault = (id: string) => {
    setAddresses((prev) =>
      prev.map((a) => ({ ...a, is_default: a.id === id }))
    )
  }

  if (loading) {
    return (
      <div className="p-4 space-y-4">
        <Skeleton animated />
        <Skeleton animated />
      </div>
    )
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {addresses.length === 0 ? (
        <Empty description="暂无收货地址" />
      ) : (
        <div className="p-2 space-y-2">
          {addresses.map((address) => (
            <SwipeAction
              key={address.id}
              rightActions={[
                {
                  key: 'delete',
                  text: <DeleteOutline />,
                  color: 'danger',
                  onClick: () => handleDelete(address.id),
                },
              ]}
            >
              <Card
                className={address.is_default ? 'border-primary-500' : ''}
                onClick={() => navigate(`/address/edit/${address.id}`)}
              >
                <div className="flex justify-between items-start">
                  <div className="flex-1">
                    <div className="flex items-center gap-2">
                      <span className="font-bold">{address.receiver}</span>
                      <span className="text-gray-600">{address.phone}</span>
                      {address.is_default && (
                        <span className="text-xs text-primary-500 bg-primary-50 px-1 rounded">
                          默认
                        </span>
                      )}
                    </div>
                    <div className="text-sm text-gray-500 mt-1">
                      {address.province}{address.city}{address.district}{address.detail}
                    </div>
                  </div>
                  <div
                    className="text-primary-500"
                    onClick={(e) => {
                      e.stopPropagation()
                      navigate(`/address/edit/${address.id}`)
                    }}
                  >
                    <EditSOutline />
                  </div>
                </div>
                {!address.is_default && (
                  <div
                    className="text-sm text-primary-500 mt-2 pt-2 border-t"
                    onClick={(e) => {
                      e.stopPropagation()
                      handleSetDefault(address.id)
                    }}
                  >
                    设为默认地址
                  </div>
                )}
              </Card>
            </SwipeAction>
          ))}
        </div>
      )}

      {/* 新增地址按钮 */}
      <div className="fixed bottom-4 left-4 right-4">
        <Button
          block
          color="primary"
          size="large"
          onClick={() => navigate('/address/edit/new')}
        >
          + 新增收货地址
        </Button>
      </div>
    </div>
  )
}