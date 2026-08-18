import { ArrowLeft } from 'lucide-react'
import type { ArticleBlock, ArticleDocument } from './document'
import { Link } from 'react-router-dom'

type ReaderProps = {
  article: ArticleDocument
}

function Block({ block }: { block: ArticleBlock }) {
  if (block.type === 'image') {
    return (
      <figure className="article-figure">
        <img src={block.src} alt={block.alt} />
        <figcaption>{block.caption}</figcaption>
      </figure>
    )
  }

  if (block.type === 'heading') {
    const Heading = block.level === 2 ? 'h2' : 'h3'
    return (
      <Heading className={`article-heading article-heading-${block.level}`} id={block.id}>
        {block.text}
      </Heading>
    )
  }

  return <p className="article-paragraph">{block.text}</p>
}

export function Reader({ article }: ReaderProps) {
  return (
    <main className="reader-shell">
      <aside className="reader-sidebar">
        <Link aria-label="Back to library" className="reader-back-button" to="/library">
          <ArrowLeft className="back-icon" strokeWidth={2.25} />
        </Link>
        <nav className="table-of-contents" aria-label="Table of contents">
          <a className="toc-entry toc-current" href="#article-title" aria-current="location">
            {article.title}
          </a>
          {article.sections.map((section) => (
            <a className="toc-entry" href={`#${section.id}`} key={section.id}>
              {section.title}
            </a>
          ))}
        </nav>
      </aside>

      <article className="article">
        <header className="article-header">
          <h1 id="article-title">{article.title}</h1>
          <p className="article-metadata">
            <span>{article.author}</span>
            <span className="metadata-separator" aria-hidden="true">
              •
            </span>
            <a href={article.source.url}>{article.source.label}</a>
          </p>
        </header>

        {article.lead.map((block, index) => (
          <Block block={block} key={`lead-${index}`} />
        ))}

        {article.sections.map((section) => (
          <section className="article-section" key={section.id}>
            <h2 className="article-heading article-heading-2" id={section.id}>
              {section.title}
            </h2>
            {section.blocks.map((block, index) => (
              <Block block={block} key={`${section.id}-${index}`} />
            ))}
          </section>
        ))}
      </article>
      <aside className="reader-right-sidebar" aria-hidden="true" />
    </main>
  )
}
