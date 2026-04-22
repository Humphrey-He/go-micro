import { type ReactNode } from 'react'

interface Props {
  children?: ReactNode
  width?: string
  height?: string
  count?: number
}

export default function LoadingSkeleton({ width = '100%', height = '16px', count = 1 }: Props) {
  return (
    <div className="space-y-2">
      {Array.from({ length: count }).map((_, i) => (
        <div
          key={i}
          className="animate-pulse bg-gray-200 rounded"
          style={{ width, height }}
        />
      ))}
    </div>
  )
}