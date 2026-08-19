import { type FormEvent, useEffect, useId, useRef, useState } from 'react'
import {
  CirclePlus,
  Library as LibraryIcon,
  PanelLeftClose,
  PanelLeftOpen,
  PanelRightOpen,
} from 'lucide-react'
import { Link } from 'react-router-dom'
import type { LibraryDocument } from './document'
import { DocumentInfoSidebar } from './DocumentInfoSidebar'

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

type ApplicationSidebarProps = Pick<LibraryProps, 'onImport'> & {
  isOpen: boolean
  onToggle: () => void
}

function ApplicationSidebar({ isOpen, onImport, onToggle }: ApplicationSidebarProps) {
  if (!isOpen) {
    return (
      <aside className="application-sidebar application-sidebar-closed">
        <div className="collapsed-application-banner">
          <button
            aria-label="Open application sidebar"
            className="sidebar-icon-button"
            onClick={onToggle}
            title="Open application sidebar"
            type="button"
          >
            <PanelLeftOpen aria-hidden="true" strokeWidth={2.25} />
          </button>
        </div>
        <nav className="collapsed-application-navigation" aria-label="Application">
          <Link aria-current="page" aria-label="Library" className="collapsed-library-navigation-item" to="/library">
            <LibraryIcon aria-hidden="true" className="library-icon" strokeWidth={2.25} />
          </Link>
        </nav>
      </aside>
    )
  }

  return (
    <aside className="application-sidebar">
      <div className="brand-banner">
        <span className="wordmark">malum</span>
        <button
          aria-label="Close application sidebar"
          className="sidebar-icon-button"
          onClick={onToggle}
          title="Close application sidebar"
          type="button"
        >
          <PanelLeftClose aria-hidden="true" strokeWidth={2.25} />
        </button>
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

function ReadyDocumentRow({
  document,
  onPreview,
}: {
  document: Extract<LibraryDocument, { status: 'ready' }>
  onPreview: () => void
}) {
  const { article } = document

  return (
    <Link
      className="library-row ready-library-row"
      onFocus={onPreview}
      onMouseEnter={onPreview}
      to={`/documents/${article.id}`}
    >
      <img className="library-thumbnail" src={article.thumbnail.src} alt={article.thumbnail.alt} />
      <div className="library-row-information">
        <h2>{article.title}</h2>
        <p className="library-description">{article.description}</p>
        <p className="library-metadata">
          <span>{article.author?.name ?? 'Unknown author'}</span>
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
  const firstReadyDocument = documents.find((document) => document.status === 'ready')
  const [applicationSidebarOpen, setApplicationSidebarOpen] = useState(true)
  const [infoSidebarOpen, setInfoSidebarOpen] = useState(Boolean(firstReadyDocument))
  const [activeDocumentId, setActiveDocumentId] = useState(firstReadyDocument?.article.id)
  const activeDocument = documents.find(
    (document) => document.status === 'ready' && document.article.id === activeDocumentId,
  )
  const activeArticle = activeDocument?.status === 'ready' ? activeDocument.article : firstReadyDocument?.article

  useEffect(() => {
    if (!activeDocumentId && firstReadyDocument) {
      setActiveDocumentId(firstReadyDocument.article.id)
      setInfoSidebarOpen(true)
    }
  }, [activeDocumentId, firstReadyDocument])

  return (
    <main
      className={`library-shell ${applicationSidebarOpen ? '' : 'application-sidebar-is-closed'} ${
        infoSidebarOpen && activeArticle ? '' : 'info-sidebar-is-closed'
      }`}
    >
      <ApplicationSidebar
        isOpen={applicationSidebarOpen}
        onImport={onImport}
        onToggle={() => setApplicationSidebarOpen((open) => !open)}
      />
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
                <ReadyDocumentRow
                  document={document}
                  key={document.article.id}
                  onPreview={() => setActiveDocumentId(document.article.id)}
                />
              ) : (
                <ImportingDocumentRow document={document} key={document.id} />
              ),
            )}
          </div>
        )}
      </section>
      {infoSidebarOpen && activeArticle ? (
        <DocumentInfoSidebar article={activeArticle} onClose={() => setInfoSidebarOpen(false)} />
      ) : activeArticle ? (
        <div className="right-sidebar-closed-control">
          <button
            aria-label="Open document information"
            className="sidebar-icon-button"
            onClick={() => setInfoSidebarOpen(true)}
            title="Open document information"
            type="button"
          >
            <PanelRightOpen aria-hidden="true" strokeWidth={2.25} />
          </button>
        </div>
      ) : null}
    </main>
  )
}
