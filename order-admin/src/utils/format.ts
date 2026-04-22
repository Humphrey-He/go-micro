import dayjs from 'dayjs'

export const formatAmount = (amount: number): string => {
  return (amount / 100).toFixed(2)
}

export const formatDateTime = (timestamp: number): string => {
  return dayjs(timestamp * 1000).format('YYYY-MM-DD HH:mm:ss')
}

export const formatDate = (timestamp: number): string => {
  return dayjs(timestamp * 1000).format('YYYY-MM-DD')
}

export const formatTime = (timestamp: number): string => {
  return dayjs(timestamp * 1000).format('HH:mm:ss')
}

export const formatDateTimeStr = (str: string): string => {
  if (!str) return '-'
  return dayjs(str).format('YYYY-MM-DD HH:mm:ss')
}

export const parseParams = (search: string): Record<string, string> => {
  const params = new URLSearchParams(search)
  const result: Record<string, string> = {}
  params.forEach((value, key) => {
    result[key] = value
  })
  return result
}
