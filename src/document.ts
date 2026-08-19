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
  level: 2 | 3
  text: string
}

export type ArticleBlock = ArticleImage | ArticleParagraph | ArticleHeading

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
  progressPercent: number
}

export type ArticleDocument = {
  id: string
  title: string
  author: Author | null
  description: string
  thumbnail: {
    src: string
    alt: string
  }
  source: {
    label: string
    url: string
  }
  details: DocumentDetails
  lead: ArticleBlock[]
  sections: ArticleSection[]
}

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
