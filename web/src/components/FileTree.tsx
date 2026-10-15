import { useCallback, useEffect, useState } from 'react'
import { ChevronRight, File, Folder, FolderOpen } from 'lucide-react'
import { listEntries } from '../api/entries'
import type { FileEntry } from '../api/entries'
import './FileTree.css'

type FileTreeProps = {
  rootPath: string
  selectedPath: string | undefined
  onOpenFile: (file: FileEntry) => void
}

const INDENT = 14

function FileTree({ rootPath, selectedPath, onOpenFile }: FileTreeProps) {
  const [childrenByPath, setChildrenByPath] = useState<Record<string, FileEntry[]>>({})
  const [expanded, setExpanded] = useState<Record<string, boolean>>({})
  const [loading, setLoading] = useState<Record<string, boolean>>({})
  const [errors, setErrors] = useState<Record<string, string>>({})

  const load = useCallback(async (path: string, signal?: AbortSignal) => {
    setLoading((current) => ({ ...current, [path]: true }))

    try {
      const entries = await listEntries(path, signal)

      setChildrenByPath((current) => ({ ...current, [path]: entries }))
      setErrors((current) => {
        const next = { ...current }

        delete next[path]

        return next
      })
    } catch (err) {
      if (err instanceof DOMException && err.name === 'AbortError') {
        return
      }

      const message = err instanceof Error ? err.message : 'failed to load folder'

      setErrors((current) => ({ ...current, [path]: message }))
    } finally {
      setLoading((current) => ({ ...current, [path]: false }))
    }
  }, [])

  useEffect(() => {
    const controller = new AbortController()

    // Loads the root listing; the state updates happen after the
    // request resolves, not synchronously.
    // eslint-disable-next-line react/set-state-in-effect
    void load(rootPath, controller.signal)

    return () => controller.abort()
  }, [load, rootPath])

  function toggleFolder(entry: FileEntry) {
    const isExpanded = expanded[entry.path] === true

    setExpanded((current) => ({ ...current, [entry.path]: !isExpanded }))

    if (!isExpanded && childrenByPath[entry.path] === undefined) {
      void load(entry.path)
    }
  }

  const renderEntries = (entries: FileEntry[], depth: number) =>
    entries.map((entry) => {
      const indent = { paddingLeft: `${8 + depth * INDENT}px` }
      const isExpanded = expanded[entry.path] === true

      if (!entry.is_dir) {
        return (
          <button
            type="button"
            className={`tree-row tree-file${selectedPath === entry.path ? ' is-selected' : ''}`}
            key={entry.path}
            style={indent}
            onClick={() => onOpenFile(entry)}
          >
            <File size={14} strokeWidth={1.75} aria-hidden="true" />
            <span className="tree-name">{entry.name}</span>
          </button>
        )
      }

      return (
        <div key={entry.path}>
          <button
            type="button"
            className="tree-row tree-folder"
            style={indent}
            aria-expanded={isExpanded}
            onClick={() => toggleFolder(entry)}
          >
            <ChevronRight
              className={`tree-chevron${isExpanded ? ' is-open' : ''}`}
              size={14}
              strokeWidth={1.75}
              aria-hidden="true"
            />

            {isExpanded ? (
              <FolderOpen size={14} strokeWidth={1.75} aria-hidden="true" />
            ) : (
              <Folder size={14} strokeWidth={1.75} aria-hidden="true" />
            )}

            <span className="tree-name">{entry.name}</span>
          </button>

          {isExpanded &&
            (loading[entry.path] === true ? (
              <p
                className="tree-info"
                style={{ paddingLeft: `${8 + (depth + 1) * INDENT}px` }}
              >
                Loading...
              </p>
            ) : errors[entry.path] !== undefined ? (
              <p
                className="tree-info tree-info-error"
                style={{ paddingLeft: `${8 + (depth + 1) * INDENT}px` }}
              >
                {errors[entry.path]}
              </p>
            ) : (
              renderEntries(childrenByPath[entry.path] ?? [], depth + 1)
            ))}
        </div>
      )
    })

  const rootEntries = childrenByPath[rootPath]

  return (
    <div className="tree">
      {rootEntries === undefined ? (
        loading[rootPath] === true ? (
          <p className="tree-info">Loading...</p>
        ) : errors[rootPath] !== undefined ? (
          <p className="tree-info tree-info-error">{errors[rootPath]}</p>
        ) : null
      ) : rootEntries.length === 0 ? (
        <p className="tree-info">Empty project.</p>
      ) : (
        renderEntries(rootEntries, 0)
      )}
    </div>
  )
}

export default FileTree
