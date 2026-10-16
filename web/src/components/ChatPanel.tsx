import { useCallback, useEffect, useRef, useState } from 'react'
import type { KeyboardEvent } from 'react'
import { MessageSquare, Plus, Send, X } from 'lucide-react'
import { listModels } from '../api/models'
import { createSession, getMessages, listSessions, sendMessage } from '../api/sessions'
import type { ChatMessage } from '../api/sessions'
import { useApiList } from '../hooks/useApiList'
import ModelPicker from './ModelPicker'
import './ChatPanel.css'

type ChatPanelProps = {
  projectId: number
  onClose: () => void
}

function formatTime(createdAt: string): string {
  const date = new Date(`${createdAt.replace(' ', 'T')}Z`)

  if (Number.isNaN(date.getTime())) {
    return ''
  }

  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
}

function ChatPanel({ projectId, onClose }: ChatPanelProps) {
  const fetchSessions = useCallback(
    (signal?: AbortSignal) => listSessions(projectId, signal),
    [projectId],
  )

  const fetchModels = useCallback(
    (signal?: AbortSignal) => listModels(undefined, signal),
    [],
  )

  const {
    items: sessions,
    setItems: setSessions,
    loading: sessionsLoading,
    error: sessionsError,
    reload: reloadSessions,
  } = useApiList(fetchSessions)

  const { items: models, error: modelsError } = useApiList(fetchModels)

  const [selectedSessionId, setSelectedSessionId] = useState<number | null>(null)
  const [model, setModel] = useState('')

  const activeSession =
    sessions.find((session) => session.id === selectedSessionId) ?? sessions[0] ?? null
  const activeSessionId = activeSession?.id ?? null

  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [messagesLoading, setMessagesLoading] = useState(false)
  const [messagesError, setMessagesError] = useState<string | null>(null)
  const [reloadKey, setReloadKey] = useState(0)

  const [draft, setDraft] = useState('')
  const [sending, setSending] = useState(false)
  const [sendError, setSendError] = useState<string | null>(null)
  const [creating, setCreating] = useState(false)

  const listRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const sessionId = activeSessionId

    if (sessionId === null) {
      return
    }

    const controller = new AbortController()

    async function load() {
      setMessagesLoading(true)

      try {
        const loaded = await getMessages(sessionId, controller.signal)

        setMessages(loaded)
        setMessagesError(null)
        setMessagesLoading(false)
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          return
        }

        setMessagesError(err instanceof Error ? err.message : 'failed to load messages')
        setMessagesLoading(false)
      }
    }

    // Loads the session history; the state updates happen after the
    // request resolves, not synchronously.
    // eslint-disable-next-line react/set-state-in-effect
    void load()

    return () => controller.abort()
  }, [activeSessionId, reloadKey])

  useEffect(() => {
    const element = listRef.current

    if (element !== null) {
      element.scrollTop = element.scrollHeight
    }
  }, [messages, sending])

  async function newChat() {
    setCreating(true)
    setSendError(null)

    try {
      const session = await createSession(projectId)

      setSessions((current) => [session, ...current])
      setSelectedSessionId(session.id)
    } catch (err) {
      setSendError(err instanceof Error ? err.message : 'failed to create chat')
    } finally {
      setCreating(false)
    }
  }

  async function send() {
    const content = draft.trim()

    if (activeSessionId === null || content === '' || sending) {
      return
    }

    setSending(true)
    setSendError(null)

    try {
      const result = await sendMessage(activeSessionId, content, model)

      setMessages((current) => [
        ...current,
        result.user_message,
        ...result.tool_messages,
        result.assistant_message,
      ])
      setDraft('')
      setSessions((current) =>
        current.map((session) =>
          session.id === result.user_message.session_id
            ? { ...session, title: result.title }
            : session,
        ),
      )
      setSending(false)
    } catch (err) {
      // The user message is stored even when the AI fails, so reload
      // the history to show it.
      setSendError(err instanceof Error ? err.message : 'failed to send message')
      setReloadKey((key) => key + 1)
      setSending(false)
    }
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if (event.key === 'Enter' && !event.shiftKey) {
      event.preventDefault()
      void send()
    }
  }

  return (
    <aside className="chat-panel" aria-label="AI chat">
      <header className="chat-header">
        <span className="chat-header-icon">
          <MessageSquare size={15} strokeWidth={1.75} aria-hidden="true" />
        </span>

        <h2 className="chat-title">AI Chat</h2>

        <button
          type="button"
          className="icon-button"
          aria-label="New chat"
          disabled={creating}
          onClick={() => void newChat()}
        >
          <Plus size={16} strokeWidth={1.75} aria-hidden="true" />
        </button>

        <button type="button" className="icon-button" aria-label="Close chat" onClick={onClose}>
          <X size={16} strokeWidth={1.75} aria-hidden="true" />
        </button>
      </header>

      {sessionsLoading && <p className="muted chat-info">Loading chats...</p>}

      {!sessionsLoading && sessionsError !== null && (
        <div className="chat-info">
          <p className="form-error" role="alert">
            {sessionsError}
          </p>
          <button type="button" className="button-secondary" onClick={reloadSessions}>
            Try again
          </button>
        </div>
      )}

      {!sessionsLoading && sessionsError === null && sessions.length === 0 && (
        <div className="chat-empty">
          <span className="chat-empty-icon">
            <MessageSquare size={20} strokeWidth={1.5} aria-hidden="true" />
          </span>

          <h3 className="chat-empty-title">Start a conversation</h3>
          <p className="chat-empty-text">
            Ask anything about this project. Every chat is saved as its own session.
          </p>

          <button
            type="button"
            className="button-primary"
            disabled={creating}
            onClick={() => void newChat()}
          >
            <Plus size={15} strokeWidth={2} aria-hidden="true" />
            New chat
          </button>
        </div>
      )}

      {!sessionsLoading && sessionsError === null && sessions.length > 0 && (
        <>
          <div className="chat-session-bar">
            <select
              className="chat-select"
              aria-label="Session"
              value={activeSessionId ?? ''}
              onChange={(event) => setSelectedSessionId(Number(event.target.value))}
            >
              {sessions.map((session) => (
                <option key={session.id} value={session.id}>
                  {session.title === '' ? 'New chat' : session.title}
                </option>
              ))}
            </select>
          </div>

          <div className="chat-messages" ref={listRef}>
            {messagesLoading && <p className="muted chat-info">Loading messages...</p>}

            {!messagesLoading && messagesError !== null && (
              <p className="form-error chat-info" role="alert">
                {messagesError}
              </p>
            )}

            {!messagesLoading && messagesError === null && messages.length === 0 && (
              <p className="muted chat-info">Ask anything about this project.</p>
            )}

            {!messagesLoading &&
              messages.map((message) => {
                if (message.role === 'tool') {
                  // Tool calls are shown as faint status lines,
                  // e.g. "Read File: src/main.go".
                  return (
                    <div className="chat-message chat-message-tool" key={message.id}>
                      <p className="chat-tool-text">
                        {message.content.split('\n')[0]}
                      </p>
                    </div>
                  )
                }

                if (message.role === 'user') {
                  return (
                    <div className="chat-message chat-message-user" key={message.id}>
                      <div className="chat-bubble-group">
                        <p className="chat-bubble">{message.content}</p>
                        <time className="chat-time" dateTime={message.created_at}>
                          {formatTime(message.created_at)}
                        </time>
                      </div>
                    </div>
                  )
                }

                return (
                  <div className="chat-message chat-message-assistant" key={message.id}>
                    <div className="chat-text-group">
                      <p className="chat-text">{message.content}</p>
                      <time className="chat-time" dateTime={message.created_at}>
                        {formatTime(message.created_at)}
                      </time>
                    </div>
                  </div>
                )
              })}

            {sending && (
              <div className="chat-message chat-message-assistant">
                <span className="chat-typing" aria-label="AI is thinking">
                  <span className="chat-dot" />
                  <span className="chat-dot" />
                  <span className="chat-dot" />
                </span>
              </div>
            )}
          </div>

          {sendError !== null && (
            <p className="form-error chat-info" role="alert">
              {sendError}
            </p>
          )}

          <div className="chat-input-area">
            <div className="chat-input-row">
              <textarea
                className="chat-input"
                value={draft}
                onChange={(event) => setDraft(event.target.value)}
                onKeyDown={handleKeyDown}
                placeholder="Send a message..."
                aria-label="Message"
                rows={1}
                spellCheck={false}
              />

              <button
                type="button"
                className="chat-send"
                aria-label="Send message"
                disabled={sending || draft.trim() === ''}
                onClick={() => void send()}
              >
                <Send size={15} strokeWidth={2} aria-hidden="true" />
              </button>
            </div>

            <div className="chat-input-foot">
              <ModelPicker
                models={models}
                value={model}
                onChange={setModel}
                error={modelsError}
              />

              <span className="chat-input-hint">Enter to send</span>
            </div>
          </div>
        </>
      )}
    </aside>
  )
}

export default ChatPanel
