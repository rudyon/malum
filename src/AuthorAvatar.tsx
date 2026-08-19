import { Avatar, Style } from '@dicebear/core'
import lineFaceDefinition from '@dicebear/styles/line-face.json'
import { useMemo } from 'react'
import type { Author } from './document'
import unknownAuthorAvatar from './assets/unknown-author.svg'

const lineFaceStyle = new Style(lineFaceDefinition)

function makeGeneratedAvatar(handle: string) {
  const avatar = new Avatar(lineFaceStyle, {
    seed: handle,
    backgroundColor: ['f4f1ea'],
    scale: 1.3,
  })

  return `data:image/svg+xml,${encodeURIComponent(avatar.toString())}`
}

export function AuthorAvatar({ author }: { author: Author | null }) {
  const generatedAvatar = useMemo(
    () => (author && !author.avatar ? makeGeneratedAvatar(author.handle) : undefined),
    [author],
  )
  const src = author?.avatar?.src ?? generatedAvatar ?? unknownAuthorAvatar

  return <img alt="" className="author-avatar" src={src} />
}
