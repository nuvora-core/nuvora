import { useCallback, useState } from 'react'
import { ArrowLeft, File, X } from 'lucide-react'
import type { FileEntry } from '../api/entries'
import FileTree from '../components/FileTree'
import FileViewer from '../components/FileViewer'
import { useProjects } from '../hooks/useProjects'
import logo from '../assets/logo.svg'
import './ProjectPage.css'

type ProjectPageProps = {
  projectId: number
  onBack: () => void
}

function ProjectPage({ projectId, onBack }: ProjectPageProps) {
  const { projects, loading, error, reload } = useProjects()
  const project = projects.find((item) => item.id === projectId)

  const [openFiles, setOpenFiles] = useState<FileEntry[]>([])
  const [activePath, setActivePath] = useState<string | null>(null)
  const [dirtyPaths, setDirtyPaths] = useState<Record<string, boolean>>({})

  const handleDirtyChange = useCallback((path: string, dirty: boolean) => {
    setDirtyPaths((current) => {
      if (current[path] === dirty) {
        return current
      }

      const next = { ...current }

      if (dirty) {
        next[path] = true
      } else {
        delete next[path]
      }

      return next
    })
  }, [])

  function openFile(file: FileEntry) {
    setOpenFiles((current) =>
      current.some((item) => item.path === file.path) ? current : [...current, file],
    )

    setActivePath(file.path)
  }

  function closeFile(path: string) {
    if (dirtyPaths[path] === true) {
      const file = openFiles.find((item) => item.path === path)
      const label = file?.name ?? 'This file'

      if (!window.confirm(`"${label}" has unsaved changes. Discard them?`)) {
        return
      }
    }

    const index = openFiles.findIndex((file) => file.path === path)
    const next = openFiles.filter((file) => file.path !== path)

    setOpenFiles(next)

    setDirtyPaths((current) => {
      const updated = { ...current }

      delete updated[path]

      return updated
    })

    if (activePath === path) {
      // Activate the tab on the right, otherwise the one on the left.
      const fallback = next[index] ?? next[index - 1] ?? null

      setActivePath(fallback?.path ?? null)
    }
  }

  return (
    <div className="workspace">
      <aside className="workspace-sidebar" aria-label="Project explorer">
        <div className="workspace-header">
          <button
            type="button"
            className="icon-button"
            aria-label="Back to projects"
            onClick={onBack}
          >
            <ArrowLeft size={16} strokeWidth={1.75} aria-hidden="true" />
          </button>

          <span className="workspace-title">
            {project?.name ?? `Project #${projectId}`}
          </span>
        </div>

        <div className="workspace-explorer">
          {loading && <p className="muted workspace-message">Loading...</p>}

          {!loading && error !== null && (
            <div className="workspace-message">
              <p className="form-error" role="alert">
                {error}
              </p>
              <button type="button" className="button-secondary" onClick={reload}>
                Try again
              </button>
            </div>
          )}

          {!loading && error === null && project === undefined && (
            <p className="muted workspace-message">Project not found.</p>
          )}

          {!loading && error === null && project !== undefined && (
            <FileTree
              rootPath={project.root_path}
              selectedPath={activePath ?? undefined}
              onOpenFile={openFile}
            />
          )}
        </div>
      </aside>

      <main className="workspace-main">
        {openFiles.length === 0 ? (
          <div className="workspace-welcome">
            <img className="workspace-watermark" src={logo} alt="" />

            {project !== undefined && (
              <div className="workspace-empty">
                <h1 className="workspace-name">{project.name}</h1>
                <p className="workspace-path">{project.root_path}</p>
              </div>
            )}
          </div>
        ) : (
          <>
            <div className="tabs" role="tablist" aria-label="Open files">
              {openFiles.map((file) => {
                const isActive = file.path === activePath

                return (
                  <div className={`tab${isActive ? ' is-active' : ''}`} key={file.path}>
                    <button
                      type="button"
                      role="tab"
                      className="tab-button"
                      aria-selected={isActive}
                      title={file.path}
                      onClick={() => setActivePath(file.path)}
                    >
                      <File size={13} strokeWidth={1.75} aria-hidden="true" />
                      <span className="tab-name">{file.name}</span>
                    </button>

                    {dirtyPaths[file.path] === true && (
                      <span className="tab-dirty" title="Unsaved changes" aria-hidden="true" />
                    )}

                    <button
                      type="button"
                      className="tab-close"
                      aria-label={`Close ${file.name}`}
                      onClick={() => closeFile(file.path)}
                    >
                      <X size={15} strokeWidth={2} aria-hidden="true" />
                    </button>
                  </div>
                )
              })}
            </div>

            {/* All viewers stay mounted so unsaved buffers survive tab switches. */}
            {openFiles.map((file) => (
              <div
                className="viewer-slot"
                key={file.path}
                hidden={file.path !== activePath}
              >
                <FileViewer path={file.path} onDirtyChange={handleDirtyChange} />
              </div>
            ))}
          </>
        )}
      </main>
    </div>
  )
}

export default ProjectPage
