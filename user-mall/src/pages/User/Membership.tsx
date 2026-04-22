import { useState } from 'react'
import { Card, Button, Badge } from 'antd-mobile'

interface Level {
  id: number
  name: string
  icon: string
  discount: string
  rights: string[]
  nextRights: string[]
  growthRequired: number
}

const levels: Level[] = [
  {
    id: 1,
    name: '普通会员',
    icon: '👤',
    discount: '无折扣',
    rights: ['基础购物', '积分获取'],
    nextRights: ['98折优惠', '免运门槛降低'],
    growthRequired: 0,
  },
  {
    id: 2,
    name: '银卡会员',
    icon: '🥈',
    discount: '98折',
    rights: ['98折优惠', '免运门槛降低', '每月1张满减券'],
    nextRights: ['95折优惠', '专属客服'],
    growthRequired: 1000,
  },
  {
    id: 3,
    name: '金卡会员',
    icon: '🥇',
    discount: '95折',
    rights: ['95折优惠', '专属客服', '每月2张满减券', '生日礼包'],
    nextRights: ['92折优惠', '优先发货'],
    growthRequired: 5000,
  },
  {
    id: 4,
    name: '铂金会员',
    icon: '💎',
    discount: '92折',
    rights: ['92折优惠', '优先发货', '专属活动', '1对1客服'],
    nextRights: ['88折优惠', '全年免运'],
    growthRequired: 20000,
  },
  {
    id: 5,
    name: '黑钻会员',
    icon: '👑',
    discount: '88折',
    rights: ['88折优惠', '全年免运', '1对1客服', '生日礼包', '专属定制服务'],
    nextRights: [],
    growthRequired: 100000,
  },
]

export default function Membership() {
  const [currentLevel] = useState(2)
  const [growthValue] = useState(3500)

  const level = levels[currentLevel - 1]
  const nextLevel = levels[currentLevel]
  const progress = nextLevel
    ? ((growthValue - level.growthRequired) / (nextLevel.growthRequired - level.growthRequired)) * 100
    : 100

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 会员卡片 */}
      <div className="bg-gradient-to-br from-amber-500 to-amber-600 text-white p-6">
        <div className="flex items-center justify-between mb-4">
          <div className="flex items-center gap-3">
            <span className="text-4xl">{level.icon}</span>
            <div>
              <div className="text-lg font-bold">{level.name}</div>
              <div className="text-sm opacity-80">当前等级</div>
            </div>
          </div>
          <div className="text-right">
            <div className="text-sm opacity-80">成长值</div>
            <div className="text-xl font-bold">{growthValue}</div>
          </div>
        </div>

        {nextLevel && (
          <div className="bg-white bg-opacity-20 rounded-lg p-3">
            <div className="flex justify-between text-sm mb-1">
              <span>距离 {nextLevel.name} 还差</span>
              <span>{nextLevel.growthRequired - growthValue} 成长值</span>
            </div>
            <div className="h-2 bg-white bg-opacity-30 rounded-full overflow-hidden">
              <div
                className="h-full bg-white rounded-full transition-all"
                style={{ width: `${Math.min(progress, 100)}%` }}
              />
            </div>
          </div>
        )}

        {!nextLevel && (
          <div className="bg-white bg-opacity-20 rounded-lg p-3 text-center">
            <div className="text-sm">您已是最高等级会员</div>
            <div className="text-xs opacity-80 mt-1">感谢您的支持</div>
          </div>
        )}
      </div>

      {/* 成长值任务 */}
      <Card className="m-2">
        <div className="font-bold mb-3">快速成长</div>
        <div className="space-y-3">
          {[
            { task: '下单支付', reward: '订单金额×1', completed: false },
            { task: '评价商品', reward: '+10/次', completed: true },
            { task: '完善用户资料', reward: '+20', completed: false },
            { task: '每日签到', reward: '+5/天', completed: true },
          ].map((item, index) => (
            <div
              key={index}
              className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0"
            >
              <div>
                <div className={item.completed ? 'line-through text-gray-400' : ''}>
                  {item.task}
                </div>
                <div className="text-xs text-primary-500">{item.reward}</div>
              </div>
              <Button size="small" disabled={item.completed}>
                {item.completed ? '已完成' : '去完成'}
              </Button>
            </div>
          ))}
        </div>
      </Card>

      {/* 等级权益 */}
      <Card className="m-2">
        <div className="font-bold mb-3">等级权益</div>
        <div className="space-y-4">
          {levels.map((l) => (
            <div
              key={l.id}
              className={`p-3 rounded-lg border-2 ${
                l.id === currentLevel
                  ? 'border-primary-500 bg-primary-50'
                  : 'border-gray-200'
              }`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-2">
                  <span className="text-2xl">{l.icon}</span>
                  <div>
                    <div className="font-bold">{l.name}</div>
                    <div className="text-xs text-primary-500">{l.discount}</div>
                  </div>
                </div>
                {l.id === currentLevel && (
                  <Badge color="var(--adm-color-primary)">当前</Badge>
                )}
                {l.id < currentLevel && (
                  <span className="text-xs text-gray-400">已解锁</span>
                )}
                {l.id > currentLevel && (
                  <span className="text-xs text-gray-400">
                    需{l.growthRequired}成长值
                  </span>
                )}
              </div>
              {l.id === currentLevel && (
                <div className="mt-2 pt-2 border-t border-primary-200">
                  <div className="text-sm text-gray-500 mb-1">当前权益</div>
                  <div className="flex flex-wrap gap-1">
                    {l.rights.map((right, i) => (
                      <span
                        key={i}
                        className="text-xs bg-white px-2 py-0.5 rounded border border-primary-200"
                      >
                        {right}
                      </span>
                    ))}
                  </div>
                </div>
              )}
            </div>
          ))}
        </div>
      </Card>

      {/* 升级条件 */}
      {nextLevel && (
        <Card className="m-2">
          <div className="font-bold mb-3">
            升级到 {nextLevel.name}
          </div>
          <div className="text-sm text-gray-500 space-y-2">
            <p>• 再获 {nextLevel.growthRequired - growthValue} 成长值即可升级</p>
            <p>• 升级后可享受 {nextLevel.discount} 优惠</p>
            <p>• 新增权益: {nextLevel.nextRights.join('、')}</p>
          </div>
          <Button block color="primary" className="mt-3">
            立即升级
          </Button>
        </Card>
      )}
    </div>
  )
}