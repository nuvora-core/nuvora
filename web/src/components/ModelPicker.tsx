import { useEffect, useMemo, useRef, useState } from 'react'
import type { KeyboardEvent } from 'react'
import { Check, ChevronDown, Search } from 'lucide-react'
import './ModelPicker.css'

type ModelPickerProps = {
  models: string[]
  value: string
  onChange: (model: string) => void
  error?: string | null
}

const DEFAULT_LABEL = 'Default model'

function ModelPicker({ models, value, onChange, error }: ModelPickerProps) {
  const [open, setOpen] = useState(false)
  const [query, setQuery] = useState('')
  const [highlighted, setHighlighted] = useState(0)

  const rootRef = useRef<HTMLDivElement>(null)
  const listRef = useRef<HTMLUListElement>(null)
  const searchRef = useRef<HTMLInputElement>(null)
  const buttonRef = useRef<HTMLButtonElement>(null)

  // An empty model means "use the provider default".
  const options = useMemo(() => ['', ...models], [models])

  const needle = query.trim().toLowerCase()
  const filtered =
    needle === ''
      ? options
      : options.filter((option) =>
          (option === '' ? DEFAULT_LABEL : option).toLowerCase().includes(needle),
        )

  useEffect(() => {
    if (open) {
      searchRef.current?.focus()
    }
  }, [open])

  useEffect(() => {
    if (!open) {
      return
    }

    function handlePointerDown(event: MouseEvent) {
      if (rootRef.current !== null && !rootRef.current.contains(event.target as Node)) {
        setOpen(false)
      }
    }

    document.addEventListener('mousedown', handlePointerDown)

    return () => document.removeEventListener('mousedown', handlePointerDown)
  }, [open])

  useEffect(() => {
    if (!open) {
      return
    }

    const item = listRef.current?.querySelector('.is-highlighted')

    if (item instanceof HTMLElement) {
      item.scrollIntoView({ block: 'nearest' })
    }
  }, [open, highlighted])

  function openPicker() {
    const index = options.indexOf(value)

    setQuery('')
    setHighlighted(index === -1 ? 0 : index)
    setOpen(true)
  }

  function select(model: string) {
    onChange(model)
    setOpen(false)
  }

  function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      setOpen(false)
      buttonRef.current?.focus()

      return
    }

    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setHighlighted((current) => Math.min(current + 1, filtered.length - 1))

      return
    }

    if (event.key === 'ArrowUp') {
      event.preventDefault()
      setHighlighted((current) => Math.max(current - 1, 0))

      return
    }

    if (event.key === 'Enter') {
      event.preventDefault()

      const selected = filtered[highlighted]

      if (selected !== undefined) {
        select(selected)
      }
    }
  }

  return (
    <div className="model-picker" ref={rootRef}>
      <button
        ref={buttonRef}
        type="button"
        className="model-picker-button"
        aria-haspopup="listbox"
        aria-expanded={open}
        title={error ?? undefined}
        onClick={() => (open ? setOpen(false) : openPicker())}
      >
        <span className="model-picker-label">
          {value === '' ? DEFAULT_LABEL : value}
        </span>

        <ChevronDown
          className={`model-picker-chevron${open ? ' is-open' : ''}`}
          size={14}
          strokeWidth={2}
          aria-hidden="true"
        />
      </button>

      {open && (
        <div className="model-picker-popup">
          <div className="model-picker-search">
            <Search size={14} strokeWidth={1.75} aria-hidden="true" />

            <input
              ref={searchRef}
              className="model-picker-search-input"
              value={query}
              onChange={(event) => {
                setQuery(event.target.value)
                setHighlighted(0)
              }}
              onKeyDown={handleKeyDown}
              placeholder="Search models..."
              aria-label="Search models"
              spellCheck={false}
            />
          </div>

          <ul className="model-picker-list" role="listbox" aria-label="Models" ref={listRef}>
            {filtered.length === 0 && (
              <li className="model-picker-empty">No models found.</li>
            )}

            {filtered.map((model, index) => (
              <li
                key={model === '' ? '__default__' : model}
                role="option"
                aria-selected={model === value}
                className={`model-picker-option${
                  index === highlighted ? ' is-highlighted' : ''
                }`}
                onMouseEnter={() => setHighlighted(index)}
                onClick={() => select(model)}
              >
                <span className="model-picker-option-name">
                  {model === '' ? DEFAULT_LABEL : model}
                </span>

                {model === value && (
                  <Check
                    className="model-picker-check"
                    size={14}
                    strokeWidth={2}
                    aria-hidden="true"
                  />
                )}
              </li>
            ))}
          </ul>
        </div>
      )}
    </div>
  )
}

export default ModelPicker
