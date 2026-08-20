import type {
  ApiArticleDocument,
  ArticleBlock,
  ArticleHeading,
  ArticleOutlineItem,
  ArticleTable,
} from './document'

type ApiAuthor = {
  id: string
  displayName: string
  handle: string
  avatarUrl?: string
}

type ApiImage = {
  url: string
  alt?: string
  caption?: string
}

type ApiBlock = {
  kind: 'paragraph' | 'heading' | 'image' | 'list' | 'definitions' | 'quote' | 'preformatted' | 'divider' | 'html'
  id?: string
  level?: number
  text?: string
  html?: string
  image?: ApiImage
  list?: {
    ordered: boolean
    items: Array<{ text: string; html?: string }>
  }
  definitions?: Array<{ term: string; description: string }>
}

type ApiDocument = {
  id: string
  readingKind: string
  acquisitionMethod: string
  originalFormat: string
  title: string
  description?: string
  source: { url: string; siteName: string }
  author: ApiAuthor | null
  language?: string
  publishedAt?: string
  sourceModifiedAt?: string
  wordCount: number
  readingTimeMinutes: number
  thumbnailUrl?: string
  savedAt: string
  content?: { blocks: ApiBlock[] | null; outline: ArticleOutlineItem[] | null }
}

type ApiErrorResponse = { error?: { code?: string; message?: string } }

export class ApiError extends Error {
  readonly status: number
  readonly code?: string

  constructor(message: string, status: number, code?: string) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
  }
}

export async function listDocuments(signal?: AbortSignal) {
  const response = await request<{ documents: ApiDocument[] }>('/api/documents', { signal })
  return response.documents.map(toArticleDocument)
}

export async function importDocument(url: string) {
  const response = await request<{
    document: ApiDocument
    warnings: Array<{ code: string; message: string }> | null
  }>(
    '/api/documents',
    { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ url }) },
  )
  return { article: toArticleDocument(response.document), warnings: response.warnings ?? [] }
}

export async function getDocument(documentId: string, signal?: AbortSignal) {
  const response = await request<{ document: ApiDocument }>(`/api/documents/${encodeURIComponent(documentId)}`, {
    signal,
  })
  return toArticleDocument(response.document)
}

async function request<T>(input: string, init?: RequestInit): Promise<T> {
  let response: Response
  try {
    response = await fetch(input, init)
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') throw error
    throw new ApiError('Malum could not reach its server.', 0)
  }
  if (!response.ok) {
    let body: ApiErrorResponse = {}
    try {
      body = (await response.json()) as ApiErrorResponse
    } catch {
      // The stable fallback below also covers a malformed server response.
    }
    throw new ApiError(body.error?.message ?? 'Malum could not complete this request.', response.status, body.error?.code)
  }
  return (await response.json()) as T
}

function toArticleDocument(document: ApiDocument): ApiArticleDocument {
  const sourceLabel = sourceHostname(document.source.url).replace(/^www\./i, '')
  return {
    id: document.id,
    title: document.title,
    author: document.author
      ? {
          id: document.author.id,
          name: document.author.displayName,
          handle: document.author.handle,
          avatar: document.author.avatarUrl
            ? { src: document.author.avatarUrl, alt: `${document.author.displayName}'s avatar` }
            : undefined,
        }
      : null,
    description: document.description ?? '',
    thumbnail: document.thumbnailUrl
      ? { src: document.thumbnailUrl, alt: `Thumbnail for ${document.title}` }
      : undefined,
    source: { label: sourceLabel, url: document.source.url },
    details: {
      type: document.readingKind,
      published: document.publishedAt ? formatDate(document.publishedAt) : undefined,
      readingTimeMinutes: document.readingTimeMinutes,
      wordCount: document.wordCount,
      saved: formatRelativeTime(document.savedAt),
    },
    content: document.content
      ? {
          blocks: (document.content.blocks ?? []).flatMap((block) => {
            const mapped = toArticleBlock(block)
            return mapped ? [mapped] : []
          }),
          outline: document.content.outline ?? [],
        }
      : undefined,
  }
}

function toArticleBlock(block: ApiBlock): ArticleBlock | null {
  switch (block.kind) {
    case 'paragraph':
      return block.text ? { type: 'paragraph', text: block.text } : null
    case 'heading':
      return block.id && block.text
        ? { type: 'heading', id: block.id, level: headingLevel(block.level), text: block.text }
        : null
    case 'image':
      return block.image
        ? { type: 'image', src: block.image.url, alt: block.image.alt ?? '', caption: block.image.caption ?? '' }
        : null
    case 'list':
      return block.list
        ? { type: 'list', ordered: block.list.ordered, items: block.list.items.map((item) => item.text) }
        : null
    case 'definitions':
      return block.definitions ? { type: 'definitions', entries: block.definitions } : null
    case 'quote':
      return block.text ? { type: 'quote', text: block.text } : null
    case 'preformatted':
      return block.text ? { type: 'preformatted', text: block.text } : null
    case 'divider':
      return { type: 'divider' }
    case 'html':
      return block.html ? tableFromHTML(block.html) : null
  }
}

function tableFromHTML(html: string): ArticleTable | null {
  const parsed = new DOMParser().parseFromString(html, 'text/html')
  const table = parsed.querySelector('table')
  if (!table) return null
  const rows = [...table.querySelectorAll('tr')]
    .map((row) =>
      [...row.querySelectorAll(':scope > th, :scope > td')].map((cell) => ({
        text: cell.textContent?.trim() ?? '',
        heading: cell.tagName === 'TH',
      })),
    )
    .filter((row) => row.length > 0)
  if (rows.length === 0) return null
  const caption = table.querySelector(':scope > caption')?.textContent?.trim()
  return { type: 'table', caption: caption || undefined, rows }
}

function headingLevel(level?: number): ArticleHeading['level'] {
  if (level && level >= 1 && level <= 6) return level as ArticleHeading['level']
  return 2
}

function sourceHostname(url: string) {
  try {
    return new URL(url).hostname
  } catch {
    return url
  }
}

function formatDate(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  return new Intl.DateTimeFormat('en', { dateStyle: 'long' }).format(date)
}

function formatRelativeTime(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) return value
  const elapsedSeconds = Math.max(0, Math.round((Date.now() - date.getTime()) / 1000))
  if (elapsedSeconds < 60) return 'just now'
  const elapsedMinutes = Math.floor(elapsedSeconds / 60)
  if (elapsedMinutes < 60) return `${elapsedMinutes} min${elapsedMinutes === 1 ? '' : 's'} ago`
  const elapsedHours = Math.floor(elapsedMinutes / 60)
  if (elapsedHours < 24) return `${elapsedHours} hour${elapsedHours === 1 ? '' : 's'} ago`
  const elapsedDays = Math.floor(elapsedHours / 24)
  return `${elapsedDays} day${elapsedDays === 1 ? '' : 's'} ago`
}
