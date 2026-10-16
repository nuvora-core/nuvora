import { toErrorMessage } from './http'

type ModelsResponse = {
  models: string[]
}

export async function listModels(
  provider?: string,
  signal?: AbortSignal,
): Promise<string[]> {
  const query = provider === undefined ? '' : `?provider=${encodeURIComponent(provider)}`
  const res = await fetch(`/api/v1/models${query}`, { signal })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return ((await res.json()) as ModelsResponse).models ?? []
}
