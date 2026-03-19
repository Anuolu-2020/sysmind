import { createContext, useContext, useState, useCallback } from 'react';

const ErrorDialogContext = createContext(null);

export function ErrorDialogProvider({ children }) {
  const [dialog, setDialog] = useState({
    isOpen: false,
    title: '',
    message: '',
    type: 'error', // 'error' | 'warning' | 'info'
  });

  const showError = useCallback((title, message) => {
    setDialog({ isOpen: true, title, message, type: 'error' });
  }, []);

  const showWarning = useCallback((title, message) => {
    setDialog({ isOpen: true, title, message, type: 'warning' });
  }, []);

  const showInfo = useCallback((title, message) => {
    setDialog({ isOpen: true, title, message, type: 'info' });
  }, []);

  const closeDialog = useCallback(() => {
    setDialog(d => ({ ...d, isOpen: false }));
  }, []);

  return (
    <ErrorDialogContext.Provider value={{ showError, showWarning, showInfo, closeDialog }}>
      {children}
      {dialog.isOpen && (
        <>
          <div className="error-dialog-backdrop" onClick={closeDialog} />
          <div className={`error-dialog error-dialog-${dialog.type}`}>
            <div className="error-dialog-header">
              <div className="error-dialog-icon">
                {dialog.type === 'error' && (
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="15" y1="9" x2="9" y2="15" />
                    <line x1="9" y1="9" x2="15" y2="15" />
                  </svg>
                )}
                {dialog.type === 'warning' && (
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" />
                    <line x1="12" y1="9" x2="12" y2="13" />
                    <line x1="12" y1="17" x2="12.01" y2="17" />
                  </svg>
                )}
                {dialog.type === 'info' && (
                  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
                    <circle cx="12" cy="12" r="10" />
                    <line x1="12" y1="16" x2="12" y2="12" />
                    <line x1="12" y1="8" x2="12.01" y2="8" />
                  </svg>
                )}
              </div>
              <div className="error-dialog-title-section">
                <span className="error-dialog-app">SysMind</span>
                <h3 className="error-dialog-title">{dialog.title}</h3>
              </div>
            </div>
            <div className="error-dialog-body">
              <p className="error-dialog-message">{dialog.message}</p>
            </div>
            <div className="error-dialog-footer">
              <button className="error-dialog-btn" onClick={closeDialog}>
                OK
              </button>
            </div>
          </div>
        </>
      )}
    </ErrorDialogContext.Provider>
  );
}

export function useErrorDialog() {
  const context = useContext(ErrorDialogContext);
  if (!context) {
    throw new Error('useErrorDialog must be used within an ErrorDialogProvider');
  }
  return context;
}
