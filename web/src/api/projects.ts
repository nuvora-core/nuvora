import { toErrorMessage } from './http'

export type Project = {
  id: number
  name: string
  root_path: string
  created_at: string
}

type ProjectsResponse = {
  projects: Project[]
}

export async function listProjects(signal?: AbortSignal): Promise<Project[]> {
  const res = await fetch('/api/v1/project', { signal })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  const data = (await res.json()) as ProjectsResponse

  return data.projects ?? []
}

export async function createProject(name: string, rootPath: string): Promise<void> {
  const res = await fetch('/api/v1/project', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ name, root_path: rootPath }),
  })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }
}
