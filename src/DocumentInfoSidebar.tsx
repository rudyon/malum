import { PanelRightClose } from 'lucide-react'
import type { ArticleDocument } from './document'
import { AuthorAvatar } from './AuthorAvatar'

type DocumentInfoSidebarProps = {
  article: ArticleDocument
  onClose: () => void
}

function MetadataEntry({ label, value }: { label: string; value: string }) {
  return (
    <div className="info-metadata-entry">
      <dt>{label}</dt>
      <dd>{value}</dd>
    </div>
  )
}

export function DocumentInfoSidebar({ article, onClose }: DocumentInfoSidebarProps) {
  const authorName = article.author?.name ?? 'Unknown author'
  const authorHandle = article.author?.handle ?? 'unknown'
  const minutesLeft =
    article.details.progressPercent === undefined
      ? undefined
      : Math.max(
          0,
          Math.round(article.details.readingTimeMinutes * (1 - article.details.progressPercent / 100)),
        )

  return (
    <aside className="document-info-sidebar" aria-label="Document information">
      <div className="info-banner">
        <button
          aria-label="Close document information"
          className="sidebar-icon-button"
          onClick={onClose}
          title="Close document information"
          type="button"
        >
          <PanelRightClose aria-hidden="true" strokeWidth={2.25} />
        </button>
        <h2>Info</h2>
      </div>

      <div className="info-document-heading">
        <h3>{article.title}</h3>
        <p>{article.source.label}</p>
      </div>

      <div className="info-author">
        <AuthorAvatar author={article.author} />
        <div>
          <p className="info-author-name">{authorName}</p>
          <p className="info-author-handle">@{authorHandle}</p>
        </div>
      </div>

      <dl className="info-metadata">
        <MetadataEntry label="Type" value={article.details.type} />
        <MetadataEntry label="Domain" value={article.source.label} />
        {article.details.published ? (
          <MetadataEntry label="Published" value={article.details.published} />
        ) : null}
        <MetadataEntry
          label="Length"
          value={`${article.details.readingTimeMinutes} mins (${article.details.wordCount} words)`}
        />
        <MetadataEntry label="Saved" value={article.details.saved} />
        {article.details.progressPercent !== undefined && minutesLeft !== undefined ? (
          <MetadataEntry
            label="Progress"
            value={`${article.details.progressPercent}% (${minutesLeft} mins left)`}
          />
        ) : null}
      </dl>
    </aside>
  )
}
