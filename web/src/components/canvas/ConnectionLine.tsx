// SVG bezier curve connection between two nodes
interface Props {
  x1: number
  y1: number
  x2: number
  y2: number
  active?: boolean
  preview?: boolean
}

export function ConnectionLine({ x1, y1, x2, y2, active, preview }: Props) {
  const dx = Math.max(Math.abs(x2 - x1) * 0.5, 50)
  const d = `M ${x1},${y1} C ${x1 + dx},${y1} ${x2 - dx},${y2} ${x2},${y2}`
  return (
    <>
      {/* Glow layer for active connections */}
      {active && (
        <path
          d={d}
          fill="none"
          stroke="#6366f1"
          strokeWidth={6}
          strokeOpacity={0.15}
          strokeLinecap="round"
        />
      )}
      <path
        d={d}
        fill="none"
        stroke={preview ? '#6366f1' : active ? '#818cf8' : '#3f3f46'}
        strokeWidth={2}
        strokeLinecap="round"
        strokeDasharray={preview ? '6 4' : 'none'}
        className={preview ? 'animate-dash' : ''}
      />
      {/* Arrow at target */}
      {!preview && (
        <circle cx={x2} cy={y2} r={3} fill={active ? '#818cf8' : '#3f3f46'} />
      )}
    </>
  )
}
