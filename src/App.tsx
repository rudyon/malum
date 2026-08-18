import { useState } from 'react'
import { Navigate, Route, Routes, useParams } from 'react-router-dom'
import type { LibraryDocument } from './document'
import { sampleArticle } from './fixture'
import { Library } from './Library'
import { Reader } from './Reader'

const initialDocuments: LibraryDocument[] = [{ status: 'ready', article: sampleArticle }]

export function App() {
  const [documents, setDocuments] = useState<LibraryDocument[]>(initialDocuments)

  const importUrl = (url: string) => {
    setDocuments((current) => [
      ...current,
      {
        id: `import-${crypto.randomUUID()}`,
        status: 'importing',
        url,
      },
    ])
  }

  const ReaderRoute = () => {
    const { documentId } = useParams()
    const document = documents.find(
      (candidate) => candidate.status === 'ready' && candidate.article.id === documentId,
    )

    if (!document || document.status !== 'ready') {
      return <Navigate replace to="/library" />
    }

    return <Reader article={document.article} />
  }

  return (
    <Routes>
      <Route path="/library" element={<Library documents={documents} onImport={importUrl} />} />
      <Route path="/documents/:documentId" element={<ReaderRoute />} />
      <Route path="*" element={<Navigate replace to="/library" />} />
    </Routes>
  )
}

