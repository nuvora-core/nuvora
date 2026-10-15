import type { ProviderMeta } from '../lib/providers'

type ProviderLogoProps = {
  provider: ProviderMeta
  size?: number
}

function ProviderLogo({ provider, size = 22 }: ProviderLogoProps) {
  return (
    <svg
      viewBox={provider.viewBox}
      width={size}
      height={size}
      fill="currentColor"
      aria-hidden="true"
    >
      <path d={provider.path} />
    </svg>
  )
}

export default ProviderLogo
