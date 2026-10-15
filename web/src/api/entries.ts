import { toErrorMessage } from './http'

export type FileEntry = {
  name: string
  path: string
  is_dir: boolean
}

type EntriesResponse = {
  path: string
  entries: FileEntry[]
}

export async function listEntries(
  path: string,
  signal?: AbortSignal,
): Promise<FileEntry[]> {
  const res = await fetch(`/api/v1/entries?path=${encodeURIComponent(path)}`, {
    signal,
  })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  const data = (await res.json()) as EntriesResponse

  return data.entries ?? []
}
