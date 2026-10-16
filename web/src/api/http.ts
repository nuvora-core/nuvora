export async function toErrorMessage(res: Response): Promise<string> {
  try {
    const data = (await res.json()) as { error?: string; reason?: string }

    if (typeof data.error === 'string' && data.error !== '') {
      if (typeof data.reason === 'string' && data.reason !== '') {
        return `${data.error}: ${data.reason}`
      }

      return data.error
    }
  } catch {
    // The response body is not JSON.
  }

  return `request failed (${res.status})`
}
