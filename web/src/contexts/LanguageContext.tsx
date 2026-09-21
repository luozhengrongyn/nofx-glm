import { createContext, useContext, useState, ReactNode } from 'react'
import type { Language } from '../i18n/translations'

interface LanguageContextType {
  language: Language
  setLanguage: (lang: Language) => void
}

const LanguageContext = createContext<LanguageContextType | undefined>(
  undefined
)

export function LanguageProvider({ children }: { children: ReactNode }) {
  // Default to Chinese, allow user to switch
  // Clear legacy forced 'en' value from old code that always wrote 'en'
  const [language, setLanguageState] = useState<Language>(() => {
    const saved = localStorage.getItem('language') as Language | null
    // Old code forced 'en' on every load; treat it as unset so we default to 'zh'
    if (!saved || saved === 'en') {
      // Check if user explicitly chose 'en' via the switcher (v2 flag)
      const chose = localStorage.getItem('lang_explicit')
      if (!chose) {
        localStorage.setItem('language', 'zh')
        return 'zh'
      }
    }
    return saved || 'zh'
  })

  const handleSetLanguage = (lang: Language) => {
    localStorage.setItem('language', lang)
    localStorage.setItem('lang_explicit', '1')
    setLanguageState(lang)
  }

  return (
    <LanguageContext.Provider
      value={{ language, setLanguage: handleSetLanguage }}
    >
      {children}
    </LanguageContext.Provider>
  )
}

export function useLanguage() {
  const context = useContext(LanguageContext)
  if (!context) {
    throw new Error('useLanguage must be used within LanguageProvider')
  }
  return context
}
