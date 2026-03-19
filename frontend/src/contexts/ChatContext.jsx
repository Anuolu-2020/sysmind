import React, { createContext, useContext } from 'react';

const ChatContext = createContext();

export function ChatProvider({ children }) {
  // This context can be extended in the future if needed
  // For now, it's a minimal wrapper that doesn't break the app structure
  const value = {};

  return (
    <ChatContext.Provider value={value}>
      {children}
    </ChatContext.Provider>
  );
}

export function useChat() {
  const context = useContext(ChatContext);
  if (!context) {
    throw new Error('useChat must be used within ChatProvider');
  }
  return context;
}
