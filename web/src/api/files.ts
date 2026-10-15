import { toErrorMessage } from './http'

export type FileContent = {
  path: string
  name: string
  size: number
  content: string
}

export async function readFile(
  path: string,
  signal?: AbortSignal,
): Promise<FileContent> {
  const res = await fetch(`/api/v1/file?path=${encodeURIComponent(path)}`, {
    signal,
  })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return (await res.json()) as FileContent
}
