import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import {
  AreaChart, Area, XAxis, YAxis, Tooltip,
  ResponsiveContainer, CartesianGrid,
} from 'recharts'
import { api } from '../api/client'
import type { PricePoint, ChartTimeframe } from '../api/types'
import styles from './PriceChart.module.css'

interface Props {
  symbol: string
}

const TIMEFRAMES: ChartTimeframe[] = ['24h', '7d', '30d']

function formatTs(ts: number, tf: ChartTimeframe): string {
  const d = new Date(ts)
  if (tf === '24h') return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
  if (tf === '7d')  return d.toLocaleDateString([], { weekday: 'short', hour: '2-digit', minute: '2-digit' })
  return d.toLocaleDateString([], { month: 'short', day: 'numeric' })
}

export default function PriceChart({ symbol }: Props) {
  const [tf, setTf] = useState<ChartTimeframe>('24h')

  const { data: points = [], isLoading } = useQuery<PricePoint[]>({
    queryKey: ['chart', symbol, tf],
    queryFn: () => api.get<PricePoint[]>(`/api/assets/${symbol}/chart?tf=${tf}`),
    staleTime: 30 * 60 * 1000, // aligned with 30-min recorder interval
  })

  const hasData = points.length > 0

  const prices = points.map(p => p.price)
  const minP = hasData ? Math.min(...prices) : 0
  const maxP = hasData ? Math.max(...prices) : 100
  const pad = (maxP - minP) * 0.05 || 1

  return (
    <div className={styles.wrapper}>
      <div className={styles.toolbar}>
        {TIMEFRAMES.map(t => (
          <button
            key={t}
            className={`${styles.tfBtn} ${tf === t ? styles.active : ''}`}
            onClick={() => setTf(t)}
          >
            {t}
          </button>
        ))}
      </div>

      {isLoading ? (
        <div className={styles.placeholder}>Loading…</div>
      ) : !hasData ? (
        <div className={styles.placeholder}>No price history yet — place the first trade!</div>
      ) : (
        <ResponsiveContainer width="100%" height={220}>
          <AreaChart data={points} margin={{ top: 4, right: 8, bottom: 0, left: 0 }}>
            <defs>
              <linearGradient id={`grad-${symbol}`} x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%"  stopColor="#58a6ff" stopOpacity={0.25} />
                <stop offset="95%" stopColor="#58a6ff" stopOpacity={0} />
              </linearGradient>
            </defs>
            <CartesianGrid strokeDasharray="3 3" stroke="#30363d" vertical={false} />
            <XAxis
              dataKey="ts"
              tickFormatter={ts => formatTs(ts as number, tf)}
              tick={{ fill: '#8b949e', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
              minTickGap={60}
            />
            <YAxis
              domain={[minP - pad, maxP + pad]}
              tick={{ fill: '#8b949e', fontSize: 11 }}
              axisLine={false}
              tickLine={false}
              width={58}
              tickFormatter={v => (v as number).toFixed(2)}
            />
            <Tooltip
              contentStyle={{
                background: '#161b22',
                border: '1px solid #30363d',
                borderRadius: 6,
                fontSize: 12,
              }}
              labelFormatter={ts => new Date(ts as number).toLocaleString()}
              formatter={(v: number) => [v.toFixed(2), 'Price']}
            />
            <Area
              type="monotone"
              dataKey="price"
              stroke="#58a6ff"
              strokeWidth={1.5}
              fill={`url(#grad-${symbol})`}
              dot={false}
              activeDot={{ r: 3, fill: '#58a6ff' }}
              isAnimationActive={false}
            />
          </AreaChart>
        </ResponsiveContainer>
      )}
    </div>
  )
}
