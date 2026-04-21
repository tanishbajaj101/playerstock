import { useQuery } from '@tanstack/react-query'
import { LineChart, Line, ResponsiveContainer } from 'recharts'
import { api } from '../api/client'
import type { PricePoint } from '../api/types'

export default function Sparkline({ symbol }: { symbol: string }) {
  const { data: points = [] } = useQuery<PricePoint[]>({
    queryKey: ['chart', symbol, '24h'],
    queryFn: () => api.get<PricePoint[]>(`/api/assets/${symbol}/chart?tf=24h`),
    staleTime: 30 * 60 * 1000,
  })

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
