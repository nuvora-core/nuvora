export type Provider = {
  name: string
  base_url: string
}

type ProvidersResponse = {
  providers: Provider[]
}

async function toErrorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string }

    if (typeof data.error === 'string' && data.error !== '') {
      return data.error
    }
  } catch {
    // The response body is not JSON.
  }

  return `request failed (${res.status})`
}

export async function listProviders(signal?: AbortSignal): Promise<Provider[]> {
  const res = await fetch('/api/v1/provider', { signal })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  const data = (await res.json()) as ProvidersResponse

  return data.providers ?? []
}

export async function connectProvider(name: string, apiKey: string): Promise<void> {
  const res = await fetch('/api/v1/provider', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, api_key: apiKey }),
  })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }
}
