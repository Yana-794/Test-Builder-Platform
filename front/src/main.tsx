import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import '@/app/styles/variables.css'
import '@/app/styles/global.css'
import '@fontsource/inter/400.css'
import '@fontsource/inter/500.css'
import '@fontsource/inter/600.css'
import '@fontsource/inter/700.css'
import { App } from './app/App.tsx'

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
)
