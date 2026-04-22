import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Button, List } from 'antd-mobile'

// 签到任务
interface Task {
  id: string
  name: string
  reward: number
  completed: boolean
}

const mockTasks: Task[] = [
  { id: '1', name: '浏览商品5分钟', reward: 5, completed: false },
  { id: '2', name: '分享商品给好友', reward: 10, completed: false },
  { id: '3', name: '下单购买任意商品', reward: 50, completed: false },
  { id: '4', name: '评价已完成的订单', reward: 20, completed: true },
]

// 签到记录
interface SignRecord {
  date: string
  reward: number
}

export default function CheckIn() {
  const navigate = useNavigate()
  const [signed, setSigned] = useState(false)
  const [totalPoints, setTotalPoints] = useState(888)
  const [todayReward] = useState(10)
  const [continuousDays, setContinuousDays] = useState(7)
  const [tasks, setTasks] = useState<Task[]>(mockTasks)
  const [signRecords, setSignRecords] = useState<SignRecord[]>([])

  useEffect(() => {
    const today = new Date().toDateString()
    const records = localStorage.getItem('sign_records')
    if (records) {
      const parsed: SignRecord[] = JSON.parse(records)
      setSignRecords(parsed)
      if (parsed.some((r) => r.date === today)) {
        setSigned(true)
      }
    }
  }, [])

  const handleSignIn = () => {
    const today = new Date().toDateString()
    const reward = todayReward + (continuousDays >= 7 ? 5 : 0)

    setSigned(true)
    setTotalPoints((prev) => prev + reward)
    setContinuousDays((prev) => prev + 1)

    const newRecord: SignRecord = { date: today, reward }
    const newRecords = [...signRecords, newRecord]
    setSignRecords(newRecords)
    localStorage.setItem('sign_records', JSON.stringify(newRecords))
  }

  const handleTaskComplete = (taskId: string) => {
    const task = tasks.find((t) => t.id === taskId)
    if (!task || task.completed) return

    setTasks((prev) =>
      prev.map((t) => (t.id === taskId ? { ...t, completed: true } : t))
    )
    setTotalPoints((prev) => prev + task.reward)
  }

  return (
    <div className="min-h-screen bg-gray-50 pb-20">
      {/* 签到卡片 */}
      <div className="bg-gradient-to-br from-primary-500 to-primary-400 text-white p-6">
        <div className="text-center">
          <div className="text-lg opacity-80">今日签到</div>
          <div className="text-4xl font-bold mt-2">+{todayReward}</div>
          <div className="text-sm opacity-80 mt-1">
            {continuousDays >= 7 ? '额外+5连续签到奖励' : `连续签到${continuousDays}天`}
          </div>
        </div>

        {/* 签到日历（本周） */}
        <div className="flex justify-around mt-6 bg-white bg-opacity-20 rounded-lg p-4">
          {['日', '一', '二', '三', '四', '五', '六'].map((day, index) => {
            const today = new Date().getDay()
            const isToday = index === today
            const isSigned = index <= today && signed
            return (
              <div key={day} className="text-center">
                <div className={`w-8 h-8 rounded-full flex items-center justify-center ${
                  isToday ? 'bg-white text-primary-500' : ''
                } ${isSigned && !isToday ? 'bg-green-500 text-white' : ''}`}>
                  {isSigned ? '✓' : ''}
                </div>
                <div className="text-xs mt-1 opacity-80">{day}</div>
              </div>
            )
          })}
        </div>

        <Button
          block
          color={signed ? 'default' : 'warning'}
          className="mt-4"
          disabled={signed}
          onClick={handleSignIn}
        >
          {signed ? '✓ 今日已签到' : '立即签到'}
        </Button>
      </div>

      {/* 积分概览 */}
      <Card className="m-2">
        <div className="flex justify-between items-center">
          <div>
            <div className="text-sm text-gray-500">我的积分</div>
            <div className="text-2xl font-bold text-primary-500">{totalPoints}</div>
          </div>
          <Button size="small" onClick={() => navigate('/coupons')}>
            积分商城
          </Button>
        </div>
      </Card>

      {/* 签到任务 */}
      <Card className="m-2">
        <div className="flex items-center gap-2 mb-3">
          <span className="text-xl">🎁</span>
          <span className="font-bold">签到任务</span>
        </div>
        <div className="space-y-3">
          {tasks.map((task) => (
            <div
              key={task.id}
              className="flex items-center justify-between py-2 border-b border-gray-100 last:border-0"
            >
              <div>
                <div className={`${task.completed ? 'line-through text-gray-400' : ''}`}>
                  {task.name}
                </div>
                <div className="text-xs text-primary-500">+{task.reward}积分</div>
              </div>
              {task.completed ? (
                <span className="text-green-500 text-sm">已完成</span>
              ) : (
                <Button size="small" onClick={() => handleTaskComplete(task.id)}>
                  去完成
                </Button>
              )}
            </div>
          ))}
        </div>
      </Card>

      {/* 签到规则 */}
      <Card className="m-2">
        <div className="font-bold mb-2">签到规则</div>
        <div className="text-sm text-gray-500 space-y-1">
          <p>• 每日签到可获得{todayReward}积分</p>
          <p>• 连续签到7天额外奖励50积分</p>
          <p>• 连续签到30天额外奖励200积分</p>
          <p>• 积分可兑换优惠券或礼品</p>
        </div>
      </Card>

      {/* 签到记录 */}
      {signRecords.length > 0 && (
        <Card className="m-2">
          <div className="font-bold mb-2">最近签到记录</div>
          <List>
            {signRecords.slice(-5).reverse().map((record, index) => (
              <List.Item key={index}>
                <div className="flex justify-between">
                  <span>{record.date}</span>
                  <span className="text-primary-500">+{record.reward}</span>
                </div>
              </List.Item>
            ))}
          </List>
        </Card>
      )}
    </div>
  )
}