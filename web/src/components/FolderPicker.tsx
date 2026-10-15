import { useCallback, useEffect, useRef, useState } from 'react'
import { ArrowUp, Folder } from 'lucide-react'
import { listFolders } from '../api/folders'
import type { FolderListing } from '../api/folders'
import './FolderPicker.css'

type FolderPickerProps = {
  open: boolean
  onSelect: (path: string) => void
  onClose: () => void
}

function FolderPicker({ open, onSelect, onClose }: FolderPickerProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  // Remembers the last visited folder, so reopening resumes there.
  const pathRef = useRef<string | undefined>(undefined)

  const [listing, setListing] = useState<FolderListing | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const load = useCallback(async (path?: string) => {
    setLoading(true)

    try {
      const loaded = await listFolders(path)

      pathRef.current = loaded.path
      setListing(loaded)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'failed to load folders')
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    const dialog = dialogRef.current

    if (dialog === null) {
      return
    }

    if (open && !dialog.open) {
      dialog.showModal()

      // Starts loading the listing; the state updates happen after the
      // request resolves, not synchronously.
      // eslint-disable-next-line react/set-state-in-effect
      void load(pathRef.current)
    } else if (!open && dialog.open) {
      dialog.close()
    }
  }, [open, load])

  return (
    <dialog
      ref={dialogRef}
      className="modal"
      aria-labelledby="folder-picker-title"
      onClose={onClose}
      onClick={(event) => {
        if (event.target === dialogRef.current) {
          dialogRef.current?.close()
        }
      }}
    >
      <div className="modal-form">
        <h2 className="modal-title" id="folder-picker-title">
          Select folder
        </h2>

        <div className="folder-picker-path">
          <button
            type="button"
            className="icon-button"
            aria-label="Up one level"
            disabled={listing === null || listing.parent === ''}
            onClick={() => void load(listing?.parent)}
          >
            <ArrowUp size={16} strokeWidth={1.75} aria-hidden="true" />
          </button>

          <span
            className="folder-picker-current"
            title={listing?.path ?? undefined}
          >
            {listing?.path ?? '...'}
          </span>
        </div>

        <div className="folder-picker-list">
          {loading && <p className="muted">Loading folders...</p>}

          {!loading && error !== null && (
            <p className="form-error" role="alert">
              {error}
            </p>
          )}

          {!loading && error === null && listing !== null && listing.folders.length === 0 && (
            <p className="muted">No subfolders here.</p>
          )}

          {!loading &&
            error === null &&
            listing?.folders.map((folder) => (
              <button
                key={folder.path}
                type="button"
                className="folder-row"
                onClick={() => void load(folder.path)}
              >
                <Folder size={16} strokeWidth={1.75} aria-hidden="true" />
                <span className="folder-row-name">{folder.name}</span>
              </button>
            ))}
        </div>

        <div className="modal-actions">
          <button
            type="button"
            className="button-secondary"
            onClick={() => dialogRef.current?.close()}
          >
            Cancel
          </button>

          <button
            type="button"
            className="button-primary"
            disabled={listing === null || loading}
            onClick={() => {
              if (listing !== null) {
                onSelect(listing.path)
              }
            }}
          >
            Select folder
          </button>
        </div>
      </div>
    </dialog>
  )
}

export default FolderPicker
