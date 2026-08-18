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

export type ArticleDocument = {
  title: string
  author: string
  source: {
    label: string
    url: string
  }
  lead: ArticleBlock[]
  sections: ArticleSection[]
}

