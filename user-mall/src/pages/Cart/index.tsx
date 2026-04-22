import { useNavigate } from 'react-router-dom'
import { Checkbox, SwipeAction, Button, Empty } from 'antd-mobile'
import { DeleteOutline } from 'antd-mobile-icons'
import { useCartStore } from '@/stores/cartStore'

export default function Cart() {
  const navigate = useNavigate()
  const {
    items,
    toggleSelect,
    selectAll,
    updateQuantity,
    removeItem,
    getSelectedItems,
    getTotalAmount,
  } = useCartStore()

  const selectedItems = getSelectedItems()
  const totalAmount = getTotalAmount()
  const allSelected = items.filter((i) => i.is_valid).every((i) => i.is_selected)
  const validItems = items.filter((i) => i.is_valid)
  const invalidItems = items.filter((i) => !i.is_valid)

  const handleCheckout = () => {
    if (selectedItems.length === 0) return
    navigate('/checkout')
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {items.length === 0 ? (
        <Empty description="购物车是空的" />
      ) : (
        <>
          {/* 商品列表 */}
          <div className="p-2 space-y-2">
            {validItems.map((item) => (
              <SwipeAction
                key={item.id}
                rightActions={[
                  {
                    key: 'delete',
                    text: <DeleteOutline />,
                    color: 'danger',
                    onClick: () => removeItem(item.id),
                  },
                ]}
              >
                <div className="bg-white rounded-lg p-3 flex items-center gap-3">
                  <Checkbox
                    checked={item.is_selected}
                    onChange={() => toggleSelect(item.id)}
                  />
                  <img
                    src={item.image}
                    alt=""
                    className="w-20 h-20 rounded object-cover bg-gray-100"
                  />
                  <div className="flex-1 min-w-0">
                    <div className="text-sm line-clamp-2">{item.title}</div>
                    <div className="text-xs text-gray-400 mt-0.5">
                      {item.attributes.join(' / ')}
                    </div>
                    <div className="flex items-center justify-between mt-1">
                      <span className="text-primary-500 font-bold">
                        ¥{(item.price / 100).toFixed(2)}
                      </span>
                      <div className="flex items-center gap-2">
                        <button
                          className="w-6 h-6 rounded border border-gray-200 flex items-center justify-center"
                          onClick={() => updateQuantity(item.id, item.quantity - 1)}
                          disabled={item.quantity <= 1}
                        >
                          -
                        </button>
                        <span className="w-8 text-center">{item.quantity}</span>
                        <button
                          className="w-6 h-6 rounded border border-gray-200 flex items-center justify-center"
                          onClick={() => updateQuantity(item.id, item.quantity + 1)}
                          disabled={item.quantity >= item.stock}
                        >
                          +
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              </SwipeAction>
            ))}

            {/* 失效商品 */}
            {invalidItems.length > 0 && (
              <div className="mt-4">
                <div className="text-sm text-gray-500 mb-2">失效商品</div>
                {invalidItems.map((item) => (
                  <div
                    key={item.id}
                    className="bg-gray-100 rounded-lg p-3 flex items-center gap-3 opacity-50"
                  >
                    <div className="w-20 h-20 rounded bg-gray-200" />
                    <div className="flex-1">
                      <div className="text-sm line-clamp-2">{item.title}</div>
                      <div className="text-xs text-red-500 mt-1">{item.invalid_reason}</div>
                    </div>
                    <button
                      className="text-sm text-primary-500"
                      onClick={() => removeItem(item.id)}
                    >
                      删除
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* 底部操作栏 */}
          <div className="fixed bottom-16 left-0 right-0 bg-white border-t p-3 flex items-center justify-between">
            <div className="flex items-center gap-2">
              <Checkbox
                checked={allSelected}
                onChange={() => selectAll(!allSelected)}
              />
              <span className="text-sm">全选</span>
            </div>
            <div className="flex items-center gap-3">
              <div className="text-right">
                <div className="text-sm text-gray-500">
                  合计（{selectedItems.length}件）
                </div>
                <div className="text-primary-500 font-bold text-lg">
                  ¥{(totalAmount / 100).toFixed(2)}
                </div>
              </div>
              <Button
                color="primary"
                disabled={selectedItems.length === 0}
                onClick={handleCheckout}
              >
                去结算
              </Button>
            </div>
          </div>
        </>
      )}
    </div>
  )
}