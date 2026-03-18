import { createContext, useContext, useState, useEffect } from 'react';

const ThemeContext = createContext();

export const THEMES = {
  // Base themes
  dark: { id: 'dark', name: 'Dark', icon: 'moon' },
  light: { id: 'light', name: 'Light', icon: 'sun' },
  system: { id: 'system', name: 'System', icon: 'monitor' },
  
  // Cool themes
  tokyoNight: { id: 'tokyo-night', name: 'Tokyo Night', icon: 'moon' },
  dracula: { id: 'dracula', name: 'Dracula', icon: 'moon' },
  nord: { id: 'nord', name: 'Nord', icon: 'moon' },
  gruvbox: { id: 'gruvbox', name: 'Gruvbox', icon: 'sun' },
  solarizedDark: { id: 'solarized-dark', name: 'Solarized Dark', icon: 'moon' },
  solarizedLight: { id: 'solarized-light', name: 'Solarized Light', icon: 'sun' },
  monokai: { id: 'monokai', name: 'Monokai', icon: 'moon' },
  oneDark: { id: 'one-dark', name: 'One Dark', icon: 'moon' },
  catppuccin: { id: 'catppuccin', name: 'Catppuccin', icon: 'moon' },
  rosePine: { id: 'rose-pine', name: 'Rose Pine', icon: 'moon' },
  cyberpunk: { id: 'cyberpunk', name: 'Cyberpunk', icon: 'moon' },
  forest: { id: 'forest', name: 'Forest', icon: 'moon' },
};

export const THEME_LIST = Object.values(THEMES);

// Font families available
export const FONTS = [
  { id: 'system', name: 'System Default', value: "-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, sans-serif" },
  { id: 'inter', name: 'Inter', value: "'Inter', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'roboto', name: 'Roboto', value: "'Roboto', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'poppins', name: 'Poppins', value: "'Poppins', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'opensans', name: 'Open Sans', value: "'Open Sans', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'lato', name: 'Lato', value: "'Lato', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'nunito', name: 'Nunito', value: "'Nunito', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'sourcesans', name: 'Source Sans 3', value: "'Source Sans 3', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'ubuntu', name: 'Ubuntu', value: "'Ubuntu', -apple-system, BlinkMacSystemFont, sans-serif" },
  { id: 'jetbrains', name: 'JetBrains Mono', value: "'JetBrains Mono', 'SF Mono', Monaco, monospace" },
  { id: 'firacode', name: 'Fira Code', value: "'Fira Code', 'SF Mono', Monaco, monospace" },
  { id: 'cascadia', name: 'Cascadia Code', value: "'Cascadia Code', 'SF Mono', Monaco, monospace" },
];

// Font sizes
export const FONT_SIZES = [
  { id: 'xs', name: 'Extra Small', value: '12px' },
  { id: 'sm', name: 'Small', value: '13px' },
  { id: 'md', name: 'Medium', value: '14px' },
  { id: 'lg', name: 'Large', value: '15px' },
  { id: 'xl', name: 'Extra Large', value: '16px' },
];

// Google Fonts URL generator
const getGoogleFontsUrl = (fontId) => {
  const fontMap = {
    inter: 'Inter:wght@400;500;600;700',
    roboto: 'Roboto:wght@400;500;700',
    poppins: 'Poppins:wght@400;500;600;700',
    opensans: 'Open+Sans:wght@400;500;600;700',
    lato: 'Lato:wght@400;700',
    nunito: 'Nunito:wght@400;500;600;700',
    sourcesans: 'Source+Sans+3:wght@400;500;600;700',
    ubuntu: 'Ubuntu:wght@400;500;700',
    jetbrains: 'JetBrains+Mono:wght@400;500;600;700',
    firacode: 'Fira+Code:wght@400;500;600;700',
    cascadia: 'Cascadia+Code:wght@400;600;700',
  };
  
  if (fontMap[fontId]) {
    return `https://fonts.googleapis.com/css2?family=${fontMap[fontId]}&display=swap`;
  }
  return null;
};

export function ThemeProvider({ children }) {
  const [theme, setTheme] = useState(() => {
    const saved = localStorage.getItem('sysmind-theme');
    return saved || 'system';
  });

  const [font, setFont] = useState(() => {
    const saved = localStorage.getItem('sysmind-font');
    return saved || 'system';
  });

  const [fontSize, setFontSize] = useState(() => {
    const saved = localStorage.getItem('sysmind-font-size');
    return saved || 'md';
  });

  const [resolvedTheme, setResolvedTheme] = useState('dark');

  useEffect(() => {
    const updateResolvedTheme = () => {
      if (theme === 'system') {
        const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
        setResolvedTheme(prefersDark ? 'dark' : 'light');
      } else {
        setResolvedTheme(theme);
      }
    };

    updateResolvedTheme();

    // Listen for system theme changes
    const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
    const handler = () => {
      if (theme === 'system') {
        updateResolvedTheme();
      }
    };
    mediaQuery.addEventListener('change', handler);

    return () => mediaQuery.removeEventListener('change', handler);
  }, [theme]);

  // Apply theme
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', resolvedTheme);
    localStorage.setItem('sysmind-theme', theme);
  }, [theme, resolvedTheme]);

  // Load and apply font
  useEffect(() => {
    const fontUrl = getGoogleFontsUrl(font);
    const existingLink = document.getElementById('sysmind-google-font');
    
    if (fontUrl) {
      if (existingLink) {
        existingLink.href = fontUrl;
      } else {
        const link = document.createElement('link');
        link.id = 'sysmind-google-font';
        link.rel = 'stylesheet';
        link.href = fontUrl;
        document.head.appendChild(link);
      }
    } else if (existingLink) {
      existingLink.remove();
    }

    const fontConfig = FONTS.find(f => f.id === font);
    if (fontConfig) {
      document.documentElement.style.setProperty('--font-family', fontConfig.value);
    }
    
    localStorage.setItem('sysmind-font', font);
  }, [font]);

  // Apply font size
  useEffect(() => {
    const sizeConfig = FONT_SIZES.find(s => s.id === fontSize);
    if (sizeConfig) {
      document.documentElement.style.setProperty('--font-size-base', sizeConfig.value);
    }
    localStorage.setItem('sysmind-font-size', fontSize);
  }, [fontSize]);

  const currentTheme = THEMES[theme] || THEMES[Object.keys(THEMES).find(k => THEMES[k].id === theme)] || THEMES.dark;
  const currentFont = FONTS.find(f => f.id === font) || FONTS[0];
  const currentFontSize = FONT_SIZES.find(s => s.id === fontSize) || FONT_SIZES[2];

  const value = {
    theme,
    setTheme,
    resolvedTheme,
    currentTheme,
    isDark: !['light', 'solarized-light', 'gruvbox'].includes(resolvedTheme),
    font,
    setFont,
    currentFont,
    fontSize,
    setFontSize,
    currentFontSize,
  };

  return (
    <ThemeContext.Provider value={value}>
      {children}
    </ThemeContext.Provider>
  );
}

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) {
    throw new Error('useTheme must be used within a ThemeProvider');
  }
  return context;
}
