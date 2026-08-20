import { useCallback, useEffect, useState } from 'react'
import { Link, Navigate, Route, Routes, useNavigate, useParams } from 'react-router-dom'
import { ApiError, getDocument, importDocument, listDocuments } from './api'
import type { ArticleDocument, LibraryDocument } from './document'
import { Library } from './Library'
import { Reader } from './Reader'

export type FailedImport = {
  id: string
  url: string
  message: string
}

function ReaderRoute() {
  const { documentId } = useParams()
  const navigate = useNavigate()
  const [article, setArticle] = useState<ArticleDocument>()
  const [error, setError] = useState<string>()

  useEffect(() => {
    if (!documentId) {
      navigate('/library', { replace: true })
      return
    }
    const controller = new AbortController()
    setArticle(undefined)
    setError(undefined)
    getDocument(documentId, controller.signal)
      .then(setArticle)
      .catch((reason: unknown) => {
        if (reason instanceof DOMException && reason.name === 'AbortError') return
        if (reason instanceof ApiError && reason.status === 404) {
          navigate('/library', { replace: true })
          return
        }
        setError(reason instanceof Error ? reason.message : 'Malum could not load this document.')
      })
    return () => controller.abort()
  }, [documentId, navigate])

  if (error) {
    return (
      <main className="route-status">
        <p>{error}</p>
        <Link className="btn" to="/library">
          Return to library
        </Link>
      </main>
    )
  }
  if (!article) {
    return (
      <main className="route-status" aria-live="polite">
        <span className="loading loading-spinner loading-md" aria-hidden="true" />
        <p>Loading document...</p>
      </main>
    )
  }
  return <Reader article={article} />
}

export function App() {
  const [documents, setDocuments] = useState<LibraryDocument[]>([])
  const [failedImports, setFailedImports] = useState<FailedImport[]>([])
  const [libraryLoading, setLibraryLoading] = useState(true)
  const [libraryError, setLibraryError] = useState<string>()
  const [loadAttempt, setLoadAttempt] = useState(0)

  useEffect(() => {
    const controller = new AbortController()
    setLibraryLoading(true)
    setLibraryError(undefined)
    listDocuments(controller.signal)
      .then((articles) => setDocuments(articles.map((article) => ({ status: 'ready', article }))))
      .catch((reason: unknown) => {
        if (reason instanceof DOMException && reason.name === 'AbortError') return
        setLibraryError(reason instanceof Error ? reason.message : 'Malum could not load the library.')
      })
      .finally(() => {
        if (!controller.signal.aborted) setLibraryLoading(false)
      })
    return () => controller.abort()
  }, [loadAttempt])

  const importUrl = useCallback(async (url: string) => {
    const importID = `import-${crypto.randomUUID()}`
    setDocuments((current) => [{ id: importID, status: 'importing', url }, ...current])
    try {
      const result = await importDocument(url)
      setDocuments((current) =>
        current.map((document) =>
          document.status === 'importing' && document.id === importID
            ? { status: 'ready', article: result.article }
            : document,
        ),
      )
    } catch (reason) {
      setDocuments((current) =>
        current.filter((document) => document.status !== 'importing' || document.id !== importID),
      )
      setFailedImports((current) => [
        ...current,
        {
          id: importID,
          url,
          message: reason instanceof Error ? reason.message : 'Malum could not import this document.',
        },
      ])
    }
  }, [])

  const retryImport = (failure: FailedImport) => {
    setFailedImports((current) => current.filter((candidate) => candidate.id !== failure.id))
    void importUrl(failure.url)
  }

  return (
    <Routes>
      <Route
        path="/library"
        element={
          <Library
            documents={documents}
            failedImports={failedImports}
            isLoading={libraryLoading}
            loadError={libraryError}
            onDismissFailure={(failureID) =>
              setFailedImports((current) => current.filter((failure) => failure.id !== failureID))
            }
            onImport={(url) => void importUrl(url)}
            onRetryFailure={retryImport}
            onRetryLoad={() => setLoadAttempt((attempt) => attempt + 1)}
          />
        }
      />
      <Route path="/documents/:documentId" element={<ReaderRoute />} />
      <Route path="*" element={<Navigate replace to="/library" />} />
    </Routes>
  )
}
