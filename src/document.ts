export type ArticleImage = {
  type: 'image'
  src: string
  alt: string
  caption: string
}

export type ArticleParagraph = {
  type: 'paragraph'
  text: string
}

export type ArticleHeading = {
  type: 'heading'
  id: string
  level: 1 | 2 | 3 | 4 | 5 | 6
  text: string
}

export type ArticleList = {
  type: 'list'
  ordered: boolean
  items: string[]
}

export type ArticleDefinitions = {
  type: 'definitions'
  entries: Array<{ term: string; description: string }>
}

export type ArticleQuote = {
  type: 'quote'
  text: string
}

export type ArticlePreformatted = {
  type: 'preformatted'
  text: string
}

export type ArticleDivider = {
  type: 'divider'
}

export type ArticleTable = {
  type: 'table'
  caption?: string
  rows: Array<Array<{ text: string; heading: boolean }>>
}

export type ArticleBlock =
  | ArticleImage
  | ArticleParagraph
  | ArticleHeading
  | ArticleList
  | ArticleDefinitions
  | ArticleQuote
  | ArticlePreformatted
  | ArticleDivider
  | ArticleTable

export type ArticleOutlineItem = {
  id: string
  level: number
  title: string
}

export type ArticleSection = {
  id: string
  title: string
  blocks: ArticleBlock[]
}

export type Author = {
  id: string
  name: string
  handle: string
  avatar?: {
    src: string
    alt: string
  }
}

export type DocumentDetails = {
  type: string
  published?: string
  readingTimeMinutes: number
  wordCount: number
  saved: string
  progressPercent?: number
}

type ArticleDocumentBase = {
  id: string
  title: string
  author: Author | null
  description: string
  thumbnail?: {
    src: string
    alt: string
  }
  source: {
    label: string
    url: string
  }
  details: DocumentDetails
}

export type ApiArticleDocument = ArticleDocumentBase & {
  content?: {
    blocks: ArticleBlock[]
    outline: ArticleOutlineItem[]
  }
}

export type FixtureArticleDocument = ArticleDocumentBase & {
  lead: ArticleBlock[]
  sections: ArticleSection[]
}

export type ArticleDocument = ApiArticleDocument | FixtureArticleDocument

export type ReadyLibraryDocument = {
  status: 'ready'
  article: ArticleDocument
}

export type ImportingLibraryDocument = {
  id: string
  status: 'importing'
  url: string
}

export type LibraryDocument = ReadyLibraryDocument | ImportingLibraryDocument

export function articleContent(article: ArticleDocument) {
  if ('lead' in article) {
    const blocks = [
      ...article.lead,
      ...article.sections.flatMap((section) => [
        {
          type: 'heading' as const,
          id: section.id,
          level: 2 as const,
          text: section.title,
        },
        ...section.blocks,
      ]),
    ]
    const outline = blocks
      .filter((block): block is ArticleHeading => block.type === 'heading')
      .map(({ id, level, text }) => ({ id, level, title: text }))
    return { blocks, outline }
  }

  return article.content ?? { blocks: [], outline: [] }
}
