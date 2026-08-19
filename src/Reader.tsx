import { useMemo, useState } from 'react'
import {
  ArrowLeft,
  ChevronDown,
  ChevronUp,
  PanelLeftClose,
  PanelLeftOpen,
  PanelRightOpen,
} from 'lucide-react'
import type { ArticleBlock, ArticleDocument } from './document'
import { Link } from 'react-router-dom'
import { DocumentInfoSidebar } from './DocumentInfoSidebar'

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
  const collapsibleSectionIds = useMemo(
    () =>
      article.sections
        .filter((section) => section.blocks.some((block) => block.type === 'heading' && block.level === 3))
        .map((section) => section.id),
    [article.sections],
  )
  const [expandedSections, setExpandedSections] = useState(() => new Set(collapsibleSectionIds))
  const [tocSidebarOpen, setTocSidebarOpen] = useState(true)
  const [infoSidebarOpen, setInfoSidebarOpen] = useState(true)
  const allExpanded = collapsibleSectionIds.every((id) => expandedSections.has(id))
  const allCollapsed = collapsibleSectionIds.every((id) => !expandedSections.has(id))

  const setSectionExpanded = (sectionId: string, expanded: boolean) => {
    setExpandedSections((current) => {
      const next = new Set(current)
      if (expanded) next.add(sectionId)
      else next.delete(sectionId)
      return next
    })
  }

  return (
    <main className="reader-shell">
      {tocSidebarOpen ? (
        <aside className="reader-sidebar">
          <div className="reader-sidebar-banner">
            <Link aria-label="Back to library" className="sidebar-icon-button" title="Back to library" to="/library">
              <ArrowLeft aria-hidden="true" strokeWidth={2.25} />
            </Link>
            <button
              aria-label="Expand entire table of contents"
              className="sidebar-icon-button"
              disabled={allExpanded}
              onClick={() => setExpandedSections(new Set(collapsibleSectionIds))}
              title="Expand entire table of contents"
              type="button"
            >
              <ChevronDown aria-hidden="true" strokeWidth={2.25} />
            </button>
            <button
              aria-label="Collapse entire table of contents"
              className="sidebar-icon-button"
              disabled={allCollapsed}
              onClick={() => setExpandedSections(new Set())}
              title="Collapse entire table of contents"
              type="button"
            >
              <ChevronUp aria-hidden="true" strokeWidth={2.25} />
            </button>
            <span className="reader-banner-spacer" />
            <button
              aria-label="Close table of contents sidebar"
              className="sidebar-icon-button"
              onClick={() => setTocSidebarOpen(false)}
              title="Close table of contents sidebar"
              type="button"
            >
              <PanelLeftClose aria-hidden="true" strokeWidth={2.25} />
            </button>
          </div>
          <nav className="table-of-contents" aria-label="Table of contents">
            <a className="toc-entry toc-current" href="#article-title" aria-current="location">
              {article.title}
            </a>
            {article.sections.map((section) => {
              const childHeadings = section.blocks.filter(
                (block): block is Extract<ArticleBlock, { type: 'heading' }> =>
                  block.type === 'heading' && block.level === 3,
              )
              const isExpanded = expandedSections.has(section.id)

              return (
                <div className="toc-section" key={section.id}>
                  <div className="toc-section-heading">
                    {childHeadings.length > 0 ? (
                      <button
                        aria-label={`${isExpanded ? 'Collapse' : 'Expand'} ${section.title}`}
                        className="toc-section-toggle"
                        onClick={() => setSectionExpanded(section.id, !isExpanded)}
                        type="button"
                      >
                        {isExpanded ? (
                          <ChevronUp aria-hidden="true" strokeWidth={2.25} />
                        ) : (
                          <ChevronDown aria-hidden="true" strokeWidth={2.25} />
                        )}
                      </button>
                    ) : null}
                    <a className="toc-entry" href={`#${section.id}`}>
                      {section.title}
                    </a>
                  </div>
                  {childHeadings.length > 0 && isExpanded ? (
                    <div className="toc-children">
                      {childHeadings.map((heading) => (
                        <a className="toc-entry toc-child-entry" href={`#${heading.id}`} key={heading.id}>
                          {heading.text}
                        </a>
                      ))}
                    </div>
                  ) : null}
                </div>
              )
            })}
          </nav>
        </aside>
      ) : (
        <div className="left-sidebar-closed-control">
          <div className="reader-collapsed-banner">
            <Link aria-label="Back to library" className="sidebar-icon-button" title="Back to library" to="/library">
              <ArrowLeft aria-hidden="true" strokeWidth={2.25} />
            </Link>
            <button
              aria-label="Open table of contents sidebar"
              className="sidebar-icon-button"
              onClick={() => setTocSidebarOpen(true)}
              title="Open table of contents sidebar"
              type="button"
            >
              <PanelLeftOpen aria-hidden="true" strokeWidth={2.25} />
            </button>
          </div>
        </div>
      )}

      <article className="article">
        <header className="article-header">
          <h1 id="article-title">{article.title}</h1>
          <p className="article-metadata">
            <span>{article.author?.name ?? 'Unknown author'}</span>
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
      {infoSidebarOpen ? (
        <DocumentInfoSidebar article={article} onClose={() => setInfoSidebarOpen(false)} />
      ) : (
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
      )}
    </main>
  )
}
