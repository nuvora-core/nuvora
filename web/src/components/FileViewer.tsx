import { useEffect, useState } from 'react'
import { readFile } from '../api/files'
import type { FileContent } from '../api/files'
import './FileViewer.css'

type FileViewerProps = {
  path: string
}

function FileViewer({ path }: FileViewerProps) {
  const [file, setFile] = useState<FileContent | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    async function load() {
      try {
        const loaded = await readFile(path, controller.signal)

        setFile(loaded)
        setError(null)
        setLoading(false)
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
  }, [path])

  return (
    <div className="file-viewer">
      {loading && <p className="muted">Loading...</p>}

      {!loading && error !== null && (
        <p className="form-error" role="alert">
          {error}
        </p>
      )}

      {!loading && error === null && file !== null && (
        <pre className="file-viewer-code">{file.content}</pre>
      )}
    </div>
  )
}

export default FileViewer
