import { toErrorMessage } from './http'

export type FolderEntry = {
  name: string
  path: string
}

export type FolderListing = {
  path: string
  parent: string
  folders: FolderEntry[]
}

export async function listFolders(
  path?: string,
  signal?: AbortSignal,
): Promise<FolderListing> {
  const query = path === undefined ? '' : `?path=${encodeURIComponent(path)}`
  const res = await fetch(`/api/v1/folders${query}`, { signal })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return (await res.json()) as FolderListing
}
