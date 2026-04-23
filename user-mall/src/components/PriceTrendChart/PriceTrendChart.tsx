import { useMemo } from 'react'

interface PricePoint {
  date: string
  price: number
}

interface Props {
  pricePoints: PricePoint[]
  width?: number
  height?: number
  currentPrice?: number
}

export default function PriceTrendChart({
  pricePoints,
  width = 300,
  height = 150,
  currentPrice
}: Props) {
  const { minPrice, maxPrice, points, path } = useMemo(() => {
    if (!pricePoints || pricePoints.length === 0) {
      return { minPrice: 0, maxPrice: 100, points: [], path: '' }
    }

    const prices = pricePoints.map(p => p.price)
    const min = Math.min(...prices)
    const max = Math.max(...prices)
    const padding = (max - min) * 0.1 || 10

    const chartWidth = width - 40
    const chartHeight = height - 30
    const chartLeft = 30
    const chartTop = 10

    const pts = pricePoints.map((point, index) => {
      const x = chartLeft + (index / (pricePoints.length - 1 || 1)) * chartWidth
      const y = chartTop + chartHeight - ((point.price - min + padding) / (max - min + padding * 2)) * chartHeight
      return { x, y, price: point.price, date: point.date }
    })

    // Generate SVG path
    let pathD = ''
    if (pts.length > 0) {
      pathD = `M ${pts[0].x} ${pts[0].y}`
      for (let i = 1; i < pts.length; i++) {
        pathD += ` L ${pts[i].x} ${pts[i].y}`
      }
    }

    return {
      minPrice: min,
      maxPrice: max,
      points: pts,
      path: pathD
    }
  }, [pricePoints, width, height])

  const formatPrice = (price: number) => `¥${(price / 100).toFixed(2)}`

  return (
    <div className="relative" style={{ width, height }}>
      <svg width={width} height={height} className="overflow-visible">
        {/* Y-axis labels */}
        <text x={25} y={15} className="text-xs fill-gray-400">{formatPrice(maxPrice)}</text>
        <text x={25} y={height - 15} className="text-xs fill-gray-400">{formatPrice(minPrice)}</text>

        {/* Grid lines */}
        {[0, 0.25, 0.5, 0.75, 1].map((_ratio, i) => {
          const y = 10 + (height - 30) * (1 - _ratio)
          return (
            <line
              key={i}
              x1={30}
              y1={y}
              x2={width - 10}
              y2={y}
              stroke="#f0f0f0"
              strokeDasharray="2,2"
            />
          )
        })}

        {/* Current price line */}
        {currentPrice && (
          <line
            x1={30}
            y1={10 + (height - 30) * (1 - (currentPrice - minPrice + (maxPrice - minPrice) * 0.1) / ((maxPrice - minPrice) * 1.2))}
            x2={width - 10}
            y2={10 + (height - 30) * (1 - (currentPrice - minPrice + (maxPrice - minPrice) * 0.1) / ((maxPrice - minPrice) * 1.2))}
            stroke="#3b82f6"
            strokeDasharray="4,2"
            strokeOpacity="0.5"
          />
        )}

        {/* Price line */}
        {points.length > 1 && (
          <>
            <path
              d={path}
              fill="none"
              stroke="#ef4444"
              strokeWidth="2"
              strokeLinecap="round"
              strokeLinejoin="round"
            />
            {/* Data points */}
            {points.map((point, index) => (
              <circle
                key={index}
                cx={point.x}
                cy={point.y}
                r={3}
                fill="#ef4444"
              />
            ))}
          </>
        )}
      </svg>

      {/* X-axis labels */}
      {points.length > 0 && (
        <div className="absolute bottom-0 left-8 right-2 flex justify-between text-xs text-gray-400">
          <span>{points[0]?.date?.slice(5) || ''}</span>
          <span>{points[points.length - 1]?.date?.slice(5) || ''}</span>
        </div>
      )}
    </div>
  )
}
