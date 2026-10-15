export type ProviderMeta = {
  id: string
  name: string
  description: string
  viewBox: string
  path: string
}

// Official Groq logo mark, taken from the full "groq" wordmark
// on groq.com.
const GROQ_PATH =
  'M165.98 342.21H0L272.4 1.5l-68.75 220.11H369.6L97.23 562.32z'

// Official OpenRouter logo mark from openrouter.ai.
const OPENROUTER_PATH =
  'M303.9475,17.19926c42.79734,0,77.48933,34.69327,77.48933,77.48933s-34.69199,77.48933-77.48933,77.48933l76.86166,76.86244c9.76367,9.76313,2.84903,26.45667-10.95697,26.45667h-220.88335c-71.32686,0-129.14889-57.82202-129.14889-129.14889S77.64197,17.19926,148.96884,17.19926h154.97866ZM148.96884,68.85881c-42.79607,0-77.48933,34.69327-77.48933,77.48933s34.69327,77.48933,77.48933,77.48933,77.48933-34.69327,77.48933-77.48933-34.69327-77.48933-77.48933-77.48933Z'

export const PROVIDER_CATALOG: ProviderMeta[] = [
  {
    id: 'groq',
    name: 'Groq',
    description: 'Fast inference on LPU hardware',
    viewBox: '0 1.5 369.6 560.82',
    path: GROQ_PATH,
  },
  {
    id: 'openrouter',
    name: 'OpenRouter',
    description: 'One API for models from many providers',
    viewBox: '19.82 17.199 365.556 258.298',
    path: OPENROUTER_PATH,
  },
]

export function getProviderMeta(id: string): ProviderMeta | undefined {
  return PROVIDER_CATALOG.find((provider) => provider.id === id)
}
