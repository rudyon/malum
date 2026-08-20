import { useMemo, useState } from 'react'
import {
  ArrowLeft,
  ChevronDown,
  ChevronUp,
  PanelLeftClose,
  PanelLeftOpen,
  PanelRightOpen,
} from 'lucide-react'
import { Link } from 'react-router-dom'
import type { ArticleBlock, ArticleDocument, ArticleOutlineItem } from './document'
import { articleContent } from './document'
import { DocumentInfoSidebar } from './DocumentInfoSidebar'

type ReaderProps = {
  article: ArticleDocument
}

function Block({ block }: { block: ArticleBlock }) {
  switch (block.type) {
    case 'image':
      return (
        <figure className="article-figure">
          <img src={block.src} alt={block.alt} />
          {block.caption ? <figcaption>{block.caption}</figcaption> : null}
        </figure>
      )
    case 'heading': {
      const visualLevel = Math.min(6, Math.max(2, block.level)) as 2 | 3 | 4 | 5 | 6
      const Heading = `h${visualLevel}` as 'h2' | 'h3' | 'h4' | 'h5' | 'h6'
      return (
        <Heading className={`article-heading article-heading-${visualLevel}`} id={block.id}>
          {block.text}
        </Heading>
      )
    }
    case 'paragraph':
      return <p className="article-paragraph">{block.text}</p>
    case 'list': {
      const List = block.ordered ? 'ol' : 'ul'
      return (
        <List className="article-list">
          {block.items.map((item, index) => (
            <li key={index}>{item}</li>
          ))}
        </List>
      )
    }
    case 'definitions':
      return (
        <dl className="article-definitions">
          {block.entries.map((entry, index) => (
            <div className="article-definition" key={`${entry.term}-${index}`}>
              <dt>{entry.term}</dt>
              <dd>{entry.description}</dd>
            </div>
          ))}
        </dl>
      )
    case 'quote':
      return <blockquote className="article-quote">{block.text}</blockquote>
    case 'preformatted':
      return <pre className="article-preformatted">{block.text}</pre>
    case 'divider':
      return <hr className="article-divider" />
    case 'table':
      return (
        <div className="article-table-wrapper">
          <table className="article-table">
            {block.caption ? <caption>{block.caption}</caption> : null}
            <tbody>
              {block.rows.map((row, rowIndex) => (
                <tr key={rowIndex}>
                  {row.map((cell, cellIndex) => {
                    const Cell = cell.heading ? 'th' : 'td'
                    return <Cell key={cellIndex}>{cell.text}</Cell>
                  })}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )
  }
}

type OutlineGroup = {
  root: ArticleOutlineItem
  children: ArticleOutlineItem[]
}

function outlineGroups(outline: ArticleOutlineItem[]) {
  if (outline.length === 0) return []
  const rootLevel = Math.min(...outline.map((item) => item.level))
  const groups: OutlineGroup[] = []
  for (const item of outline) {
    if (item.level === rootLevel || groups.length === 0) {
      groups.push({ root: item, children: [] })
    } else {
      groups.at(-1)?.children.push(item)
    }
  }
  return groups
}

export function Reader({ article }: ReaderProps) {
  const content = useMemo(() => articleContent(article), [article])
  const tocGroups = useMemo(() => outlineGroups(content.outline), [content.outline])
  const collapsibleSectionIds = useMemo(
    () => tocGroups.filter((group) => group.children.length > 0).map((group) => group.root.id),
    [tocGroups],
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
            {tocGroups.map((group) => {
              const isExpanded = expandedSections.has(group.root.id)
              return (
                <div className="toc-section" key={group.root.id}>
                  <div className="toc-section-heading">
                    {group.children.length > 0 ? (
                      <button
                        aria-label={`${isExpanded ? 'Collapse' : 'Expand'} ${group.root.title}`}
                        className="toc-section-toggle"
                        onClick={() => setSectionExpanded(group.root.id, !isExpanded)}
                        type="button"
                      >
                        {isExpanded ? (
                          <ChevronUp aria-hidden="true" strokeWidth={2.25} />
                        ) : (
                          <ChevronDown aria-hidden="true" strokeWidth={2.25} />
                        )}
                      </button>
                    ) : null}
                    <a className="toc-entry" href={`#${group.root.id}`}>
                      {group.root.title}
                    </a>
                  </div>
                  {group.children.length > 0 && isExpanded ? (
                    <div className="toc-children">
                      {group.children.map((heading) => (
                        <a className="toc-entry toc-child-entry" href={`#${heading.id}`} key={heading.id}>
                          {heading.title}
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

        {content.blocks.map((block, index) => (
          <Block block={block} key={block.type === 'heading' ? block.id : `${block.type}-${index}`} />
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
