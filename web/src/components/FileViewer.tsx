import { useEffect, useState } from 'react'
import type { ChangeEvent, KeyboardEvent } from 'react'
import { readFile, saveFile } from '../api/files'
import type { FileContent } from '../api/files'
import './FileViewer.css'

type FileViewerProps = {
  path: string
  onDirtyChange: (path: string, dirty: boolean) => void
}

function FileViewer({ path, onDirtyChange }: FileViewerProps) {
  const [file, setFile] = useState<FileContent | null>(null)
  const [text, setText] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)
  const [saveError, setSaveError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    async function load() {
      try {
        const loaded = await readFile(path, controller.signal)

        setFile(loaded)
        setText(loaded.content)
        setError(null)
        setLoading(false)
        onDirtyChange(path, false)
      } catch (err) {
        if (err instanceof DOMException && err.name === 'AbortError') {
          return
        }

        setError(err instanceof Error ? err.message : 'failed to open file')
        setLoading(false)
      }
    }

    // Loads the file content; the state updates happen after the
    // request resolves, not synchronously.
    // eslint-disable-next-line react/set-state-in-effect
    void load()

    return () => controller.abort()
  }, [path, onDirtyChange])

  const dirty = file !== null && text !== file.content

  async function save() {
    if (file === null || saving) {
      return
    }

    setSaving(true)
    setSaveError(null)

    try {
      await saveFile(path, text)
      setFile((current) => (current === null ? current : { ...current, content: text }))
      onDirtyChange(path, false)
      setSaving(false)
    } catch (err) {
      setSaveError(err instanceof Error ? err.message : 'failed to save file')
      setSaving(false)
    }
  }

  function handleChange(event: ChangeEvent<HTMLTextAreaElement>) {
    const next = event.target.value

    setText(next)

    if (file !== null) {
      onDirtyChange(path, next !== file.content)
    }
  }

  function handleKeyDown(event: KeyboardEvent<HTMLTextAreaElement>) {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 's') {
      event.preventDefault()
      void save()
    }
  }

  return (
    <div className="file-viewer">
      {loading && <p className="muted file-viewer-message">Loading...</p>}

      {!loading && error !== null && (
        <p className="form-error file-viewer-message" role="alert">
          {error}
        </p>
      )}

      {!loading && error === null && file !== null && (
        <>
          <textarea
            className="file-editor"
            value={text}
            onChange={handleChange}
            onKeyDown={handleKeyDown}
            aria-label={`Edit ${file.name}`}
            wrap="off"
            spellCheck={false}
          />

          <div className="file-status">
            {saveError !== null ? (
              <span className="file-status-error" role="alert">
                {saveError}
              </span>
            ) : (
              <span>{saving ? 'Saving...' : dirty ? 'Unsaved changes' : 'Saved'}</span>
            )}
          </div>
        </>
      )}
    </div>
  )
}

export default FileViewer
