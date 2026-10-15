import { useRef, useState } from 'react'
import type { FormEvent } from 'react'
import { FolderKanban, FolderOpen, Plus, Search } from 'lucide-react'
import { createProject } from '../api/projects'
import FolderPicker from '../components/FolderPicker'
import { useProjects } from '../hooks/useProjects'
import './Projects.css'

function formatCreatedAt(createdAt: string): string {
  // SQLite stores "YYYY-MM-DD HH:MM:SS"; only the date is shown.
  return createdAt.length >= 10 ? createdAt.slice(0, 10) : createdAt
}

function Projects() {
  const { projects, loading, error, reload, refresh } = useProjects()

  const dialogRef = useRef<HTMLDialogElement>(null)

  const [query, setQuery] = useState('')
  const [name, setName] = useState('')
  const [rootPath, setRootPath] = useState('')
  const [creating, setCreating] = useState(false)
  const [createError, setCreateError] = useState<string | null>(null)
  const [pickerOpen, setPickerOpen] = useState(false)

  function openForm() {
    dialogRef.current?.showModal()
  }

  function closeForm() {
    dialogRef.current?.close()
  }

  // Fires on closeForm and on Escape; resets the form.
  function handleDialogClose() {
    setName('')
    setRootPath('')
    setCreateError(null)
  }

  async function handleCreate(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    const trimmedName = name.trim()
    const trimmedPath = rootPath.trim()

    if (trimmedName === '' || trimmedPath === '') {
      setCreateError('Name and root path are required')

      return
    }

    setCreating(true)
    setCreateError(null)

    try {
      await createProject(trimmedName, trimmedPath)
      closeForm()
      refresh()
    } catch (err) {
      setCreateError(err instanceof Error ? err.message : 'failed to create project')
    } finally {
      setCreating(false)
    }
  }

  const needle = query.trim().toLowerCase()
  const visible =
    needle === ''
      ? projects
      : projects.filter(
          (project) =>
            project.name.toLowerCase().includes(needle) ||
            project.root_path.toLowerCase().includes(needle),
        )

  return (
    <div className="projects">
      <div className="projects-toolbar">
        <div className="search">
          <Search
            className="search-icon"
            size={16}
            strokeWidth={1.75}
            aria-hidden="true"
          />
          <input
            className="input"
            type="search"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder="Search projects..."
            aria-label="Search projects"
          />
        </div>

        <button type="button" className="button-primary" onClick={openForm}>
          <Plus size={16} strokeWidth={2} aria-hidden="true" />
          New Project
        </button>
      </div>

      <dialog
        ref={dialogRef}
        className="modal"
        aria-labelledby="new-project-title"
        onClose={handleDialogClose}
        onClick={(event) => {
          if (event.target === dialogRef.current) {
            closeForm()
          }
        }}
      >
        <form className="modal-form" onSubmit={handleCreate}>
          <h2 className="modal-title" id="new-project-title">
            New project
          </h2>

          <label className="field">
            <span className="field-label">Name</span>
            <input
              className="input"
              value={name}
              onChange={(event) => setName(event.target.value)}
              placeholder="my-project"
              maxLength={75}
            />
          </label>

          <label className="field">
            <span className="field-label">Root path</span>
            <div className="field-row">
              <input
                className="input"
                value={rootPath}
                onChange={(event) => setRootPath(event.target.value)}
                placeholder="/home/user/code/my-project"
                spellCheck={false}
              />

              <button
                type="button"
                className="button-secondary"
                onClick={() => setPickerOpen(true)}
              >
                <FolderOpen size={16} strokeWidth={1.75} aria-hidden="true" />
                Browse
              </button>
            </div>
          </label>

          {createError !== null && (
            <p className="form-error" role="alert">
              {createError}
            </p>
          )}

          <div className="modal-actions">
            <button type="button" className="button-secondary" onClick={closeForm}>
              Cancel
            </button>

            <button type="submit" className="button-primary" disabled={creating}>
              {creating ? 'Creating...' : 'Create'}
            </button>
          </div>
        </form>
      </dialog>

      <FolderPicker
        open={pickerOpen}
        onClose={() => setPickerOpen(false)}
        onSelect={(path) => {
          setRootPath(path)
          setPickerOpen(false)
        }}
      />

      {loading && <p className="muted">Loading projects...</p>}

      {!loading && error !== null && (
        <div className="error-panel" role="alert">
          <p className="error-panel-message">{error}</p>
          <button type="button" className="button-secondary" onClick={reload}>
            Try again
          </button>
        </div>
      )}

      {!loading && error === null && projects.length === 0 && (
        <p className="muted">
          No projects yet. Create your first project with New Project.
        </p>
      )}

      {!loading && error === null && projects.length > 0 && visible.length === 0 && (
        <p className="muted">No projects match your search.</p>
      )}

      {!loading && error === null && visible.length > 0 && (
        <div className="project-grid">
          {visible.map((project) => (
            <a
              className="project-card"
              href={`#/project/${project.id}`}
              key={project.id}
            >
              <div className="project-card-header">
                <span className="project-tile">
                  <FolderKanban size={22} strokeWidth={1.75} aria-hidden="true" />
                </span>

                <div className="project-card-info">
                  <h3 className="project-name">{project.name}</h3>
                  <p className="project-path">{project.root_path}</p>
                </div>
              </div>

              <p className="project-meta">
                Created {formatCreatedAt(project.created_at)}
              </p>
            </a>
          ))}
        </div>
      )}
    </div>
  )
}

export default Projects
