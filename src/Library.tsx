import { type FormEvent, useId, useRef, useState } from 'react'
import { CirclePlus, Library as LibraryIcon } from 'lucide-react'
import { Link } from 'react-router-dom'
import type { LibraryDocument } from './document'

type LibraryProps = {
  documents: LibraryDocument[]
  onImport: (url: string) => void
}

type AddDocumentControlProps = {
  kind: 'icon' | 'button'
  onImport: (url: string) => void
}

function AddDocumentControl({ kind, onImport }: AddDocumentControlProps) {
  const dialogRef = useRef<HTMLDialogElement>(null)
  const dropdownRef = useRef<HTMLDetailsElement>(null)
  const inputId = useId()
  const [url, setUrl] = useState('')

  const openUrlDialog = () => {
    dropdownRef.current?.removeAttribute('open')
    dialogRef.current?.showModal()
  }

  const submitUrl = (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault()
    onImport(url)
    setUrl('')
    dialogRef.current?.close()
  }

  return (
    <>
      <details className={`dropdown ${kind === 'icon' ? 'dropdown-end' : ''}`} ref={dropdownRef}>
        <summary
          aria-label={kind === 'icon' ? 'Add document' : undefined}
          className={kind === 'icon' ? 'add-icon-button' : 'btn empty-add-button'}
        >
          {kind === 'icon' ? <CirclePlus className="add-icon" strokeWidth={2.25} /> : 'Add document'}
        </summary>
        <ul className="menu dropdown-content add-menu" role="menu">
          <li>
            <button onClick={openUrlDialog} type="button">
              URL
            </button>
          </li>
        </ul>
      </details>

      <dialog className="modal add-dialog" ref={dialogRef}>
        <div className="modal-box">
          <h2>Add from URL</h2>
          <form onSubmit={submitUrl}>
            <label className="fieldset" htmlFor={inputId}>
              <span className="label">URL</span>
              <input
                autoComplete="url"
                className="input w-full"
                id={inputId}
                onChange={(event) => setUrl(event.target.value)}
                placeholder="https://example.com/article"
                required
                type="url"
                value={url}
              />
            </label>
            <div className="modal-action">
              <button className="btn btn-ghost" onClick={() => dialogRef.current?.close()} type="button">
                Cancel
              </button>
              <button className="btn add-submit-button" type="submit">
                Add
              </button>
            </div>
          </form>
        </div>
        <form className="modal-backdrop" method="dialog">
          <button aria-label="Close add document dialog">close</button>
        </form>
      </dialog>
    </>
  )
}

function ApplicationSidebar({ onImport }: Pick<LibraryProps, 'onImport'>) {
  return (
    <aside className="application-sidebar">
      <div className="brand-banner">
        <span className="wordmark">malum</span>
        <AddDocumentControl kind="icon" onImport={onImport} />
      </div>
      <nav className="application-navigation" aria-label="Application">
        <Link aria-current="page" className="library-navigation-item" to="/library">
          <LibraryIcon className="library-icon" strokeWidth={2.25} />
          <span>Library</span>
        </Link>
      </nav>
    </aside>
  )
}

function displayUrl(url: string) {
  try {
    const parsed = new URL(url)
    return `${parsed.host}${parsed.pathname === '/' ? '' : parsed.pathname}`
  } catch {
    return url
  }
}

function ReadyDocumentRow({ document }: { document: Extract<LibraryDocument, { status: 'ready' }> }) {
  const { article } = document

  return (
    <Link className="library-row ready-library-row" to={`/documents/${article.id}`}>
      <img className="library-thumbnail" src={article.thumbnail.src} alt={article.thumbnail.alt} />
      <div className="library-row-information">
        <h2>{article.title}</h2>
        <p className="library-description">{article.description}</p>
        <p className="library-metadata">
          <span>{article.author}</span>
          <span aria-hidden="true">·</span>
          <span>{article.source.label}</span>
        </p>
      </div>
    </Link>
  )
}

function ImportingDocumentRow({ document }: { document: Extract<LibraryDocument, { status: 'importing' }> }) {
  return (
    <div className="library-row importing-library-row" aria-live="polite">
      <div className="library-row-information">
        <h2>{displayUrl(document.url)}</h2>
        <p className="importing-status">
          <span className="loading loading-spinner loading-sm" aria-hidden="true" />
          Importing...
        </p>
      </div>
    </div>
  )
}

export function Library({ documents, onImport }: LibraryProps) {
  return (
    <main className="library-shell">
      <ApplicationSidebar onImport={onImport} />
      <section className="library-content" aria-label="Library documents">
        {documents.length === 0 ? (
          <div className="empty-library">
            <p>
              No documents yet.
              <br />
              Add something you want to read.
            </p>
            <AddDocumentControl kind="button" onImport={onImport} />
          </div>
        ) : (
          <div className="library-list">
            {documents.map((document) =>
              document.status === 'ready' ? (
                <ReadyDocumentRow document={document} key={document.article.id} />
              ) : (
                <ImportingDocumentRow document={document} key={document.id} />
              ),
            )}
          </div>
        )}
      </section>
    </main>
  )
}
