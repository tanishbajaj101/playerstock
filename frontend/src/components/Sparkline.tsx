import { LineChart, Line, ResponsiveContainer } from 'recharts'
import type { PricePoint } from '../api/types'

export default function Sparkline({ points = [] }: { points?: PricePoint[] }) {
  if (points.length < 2) return null

  const up = points[points.length - 1].price >= points[0].price

  return (
    <ResponsiveContainer width="100%" height={40}>
      <LineChart data={points}>
        <Line
          type="monotone"
          dataKey="price"
          stroke={up ? 'var(--green)' : 'var(--red)'}
          strokeWidth={1.5}
          dot={false}
          isAnimationActive={false}
        />
      </LineChart>
    </ResponsiveContainer>
  )
}
