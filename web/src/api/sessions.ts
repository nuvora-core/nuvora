import { toErrorMessage } from './http'

export type Session = {
  id: number
  project_id: number
  title: string
  created_at: string
}

export type ChatMessage = {
  id: number
  session_id: number
  role: string
  content: string
  created_at: string
}

type SessionsResponse = {
  sessions: Session[]
}

type SessionResponse = {
  session: Session
}

type MessagesResponse = {
  session: Session
  messages: ChatMessage[]
}

export type SendMessageResponse = {
  title: string
  user_message: ChatMessage
  tool_messages: ChatMessage[]
  assistant_message: ChatMessage
}

export async function listSessions(
  projectId: number,
  signal?: AbortSignal,
): Promise<Session[]> {
  const res = await fetch(`/api/v1/session?project_id=${projectId}`, { signal })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return ((await res.json()) as SessionsResponse).sessions ?? []
}

export async function createSession(projectId: number): Promise<Session> {
  const res = await fetch('/api/v1/session', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ project_id: projectId }),
  })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return ((await res.json()) as SessionResponse).session
}

export async function getMessages(
  sessionId: number,
  signal?: AbortSignal,
): Promise<ChatMessage[]> {
  const res = await fetch(`/api/v1/session/${sessionId}/messages`, { signal })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return ((await res.json()) as MessagesResponse).messages ?? []
}

export async function sendMessage(
  sessionId: number,
  content: string,
  model: string,
): Promise<SendMessageResponse> {
  const res = await fetch(`/api/v1/session/${sessionId}/message`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ content, model }),
  })

  if (!res.ok) {
    throw new Error(await toErrorMessage(res))
  }

  return (await res.json()) as SendMessageResponse
}
