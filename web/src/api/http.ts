export async function toErrorMessage(res: Response): Promise<string> {
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
