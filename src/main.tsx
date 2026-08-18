import '@fontsource/cabin/400.css'
import '@fontsource/cabin/400-italic.css'
import '@fontsource/cabin/700.css'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { Reader } from './Reader'
import { sampleArticle } from './fixture'
import './styles.css'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Reader article={sampleArticle} />
  </StrictMode>,
)

