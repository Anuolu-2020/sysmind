import { useState, useRef, useEffect, useCallback } from 'react';

function Chat({ isConfigured = false, onNavigateToSettings }) {
  const [sessions, setSessions] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [currentSession, setCurrentSession] = useState(null);
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [showSidebar, setShowSidebar] = useState(true);
  const [showExportMenu, setShowExportMenu] = useState(false);
  const [promptTemplates, setPromptTemplates] = useState([]);
  const [streamingMessage, setStreamingMessage] = useState('');
  const [useStreaming, setUseStreaming] = useState(true);
  const [isCreatingSession, setIsCreatingSession] = useState(false);
  const messagesEndRef = useRef(null);
  const hasInitialized = useRef(false);

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [currentSession?.messages]);

  // Load prompt templates
  useEffect(() => {
    const loadTemplates = async () => {
      try {
        if (window.go?.main?.App?.GetPromptTemplates) {
          const templates = await window.go.main.App.GetPromptTemplates();
          setPromptTemplates(templates || []);
        }
      } catch (err) {
        console.error('Error loading templates:', err);
      }
    };
    loadTemplates();
  }, []);

  // Load all sessions on mount
  const loadSessions = useCallback(async () => {
    try {
      if (window.go?.main?.App?.GetAllChatSessions) {
        const allSessions = await window.go.main.App.GetAllChatSessions();
        setSessions(allSessions || []);
      }
    } catch (err) {
      console.error('Error loading sessions:', err);
    }
  }, []);

  // Load a specific session
  const loadSession = useCallback(async (sessionId) => {
    try {
      if (window.go?.main?.App?.GetChatSession) {
        const session = await window.go.main.App.GetChatSession(sessionId);
        if (session) {
          setCurrentSession(session);
          setCurrentSessionId(sessionId);
        }
      }
    } catch (err) {
      console.error('Error loading session:', err);
    }
  }, []);

  // Create new session
  const createNewSession = useCallback(async () => {
    if (isCreatingSession) return null; // Prevent multiple simultaneous creations
    
    try {
      setIsCreatingSession(true);
      if (window.go?.main?.App?.CreateChatSession) {
        const session = await window.go.main.App.CreateChatSession('');
        if (session) {
          setCurrentSession(session);
          setCurrentSessionId(session.id);
          await loadSessions();
          return session;
        }
      }
    } catch (err) {
      console.error('Error creating session:', err);
    } finally {
      setIsCreatingSession(false);
    }
    return null;
  }, [loadSessions, isCreatingSession]);

  // Initialize: load sessions on mount only
  useEffect(() => {
    if (!hasInitialized.current) {
      hasInitialized.current = true;
      loadSessions();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  // Setup streaming event listeners
  useEffect(() => {
    if (!window.runtime) return;

    const handleStreamChunk = (data) => {
      if (data.sessionID === currentSessionId) {
        setStreamingMessage(data.fullText);
      }
    };

    const handleStreamComplete = async (data) => {
      if (data.sessionID === currentSessionId) {
        setStreamingMessage('');
        setIsLoading(false);
        await loadSession(currentSessionId);
        await loadSessions();
      }
    };

    const handleStreamError = (data) => {
      if (data.sessionID === currentSessionId) {
        setStreamingMessage('');
        setIsLoading(false);
        alert('Streaming error: ' + data.error);
      }
    };

    window.runtime.EventsOn('chat:stream:chunk', handleStreamChunk);
    window.runtime.EventsOn('chat:stream:complete', handleStreamComplete);
    window.runtime.EventsOn('chat:stream:error', handleStreamError);

    return () => {
      if (window.runtime) {
        window.runtime.EventsOff('chat:stream:chunk');
        window.runtime.EventsOff('chat:stream:complete');
        window.runtime.EventsOff('chat:stream:error');
      }
    };
  }, [currentSessionId, loadSession, loadSessions]);

  // Auto-select session after sessions load, but don't auto-create
  useEffect(() => {
    if (!hasInitialized.current || isCreatingSession) return; 
    
    if (!currentSessionId && sessions.length > 0) {
      // Only auto-select if there are existing sessions
      loadSession(sessions[0].id);
    }
    // Note: Removed auto-creation of sessions - they should only be created when user sends first message
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sessions.length, hasInitialized.current, isCreatingSession]);

  const sendMessage = async (text) => {
    const question = text || input.trim();
    if (!question || isLoading) return;

    let sessionId = currentSessionId;

    // Create a session if none exists
    if (!sessionId) {
      const session = await createNewSession();
      if (!session) {
        console.error('Failed to create session');
        return;
      }
      sessionId = session.id;
    }

    setInput('');
    setIsLoading(true);
    setStreamingMessage('');

    // Optimistically add user message to UI
    const tempUserMsg = {
      id: 'temp-user',
      role: 'user',
      content: question,
      timestamp: Date.now(),
    };
    setCurrentSession(prev => ({
      ...prev,
      messages: [...(prev?.messages || []), tempUserMsg],
    }));

    try {
      // Refresh system context first
      if (window.go?.main?.App?.GetSystemContext) {
        await window.go.main.App.GetSystemContext();
      }

      // Use streaming or regular message sending
      if (useStreaming && window.go?.main?.App?.SendChatMessageStreaming) {
        await window.go.main.App.SendChatMessageStreaming(sessionId, question);
        // Streaming will handle updates via events
      } else if (window.go?.main?.App?.SendChatMessage) {
        await window.go.main.App.SendChatMessage(sessionId, question);
        // Reload the session to get updated messages
        await loadSession(sessionId);
        await loadSessions(); // Update session list (for title changes)
        setIsLoading(false);
      }
    } catch (err) {
      console.error('Error sending message:', err);
      // Add error message
      setCurrentSession(prev => ({
        ...prev,
        messages: [...(prev?.messages || []).filter(m => m.id !== 'temp-user'), {
          id: 'error-' + Date.now(),
          role: 'error',
          content: err.message || 'An error occurred',
          timestamp: Date.now(),
        }],
      }));
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      sendMessage();
    }
  };

  const deleteSession = async (sessionId, e) => {
    e.stopPropagation();
    if (!confirm('Delete this chat session?')) return;

    try {
      if (window.go?.main?.App?.DeleteChatSession) {
        await window.go.main.App.DeleteChatSession(sessionId);
        await loadSessions();
        
        if (currentSessionId === sessionId) {
          setCurrentSessionId(null);
          setCurrentSession(null);
        }
      }
    } catch (err) {
      console.error('Error deleting session:', err);
    }
  };

  const exportSession = async (format) => {
    if (!currentSessionId) return;
    
    try {
      if (window.go?.main?.App?.ExportChatSession) {
        const result = await window.go.main.App.ExportChatSession(currentSessionId, format);
        if (result.success) {
          // Create download
          const blob = new Blob([result.content], { 
            type: format === 'markdown' ? 'text/markdown' : 'application/json' 
          });
          const url = URL.createObjectURL(blob);
          const a = document.createElement('a');
          a.href = url;
          a.download = result.filename;
          document.body.appendChild(a);
          a.click();
          document.body.removeChild(a);
          URL.revokeObjectURL(url);
        } else {
          alert('Export failed: ' + result.error);
        }
      }
    } catch (err) {
      alert('Export error: ' + err.message);
    }
    setShowExportMenu(false);
  };

  const formatDate = (timestamp) => {
    const date = new Date(timestamp);
    const now = new Date();
    const diffDays = Math.floor((now - date) / (1000 * 60 * 60 * 24));

    if (diffDays === 0) {
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } else if (diffDays === 1) {
      return 'Yesterday';
    } else if (diffDays < 7) {
      return date.toLocaleDateString([], { weekday: 'short' });
    } else {
      return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
    }
  };

  const formatFullDate = (timestamp) => {
    return new Date(timestamp).toLocaleString([], {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit',
    });
  };

  const messages = currentSession?.messages || [];

  // Group templates by category
  const templatesByCategory = promptTemplates.reduce((acc, t) => {
    if (!acc[t.category]) acc[t.category] = [];
    acc[t.category].push(t);
    return acc;
  }, {});

  return (
    <div className="chat-container">
      {/* Sidebar */}
      <div className={`chat-sidebar ${showSidebar ? '' : 'collapsed'}`}>
        <div className="sidebar-header">
          <h3>Chat History</h3>
          <button className="new-chat-btn" onClick={createNewSession} title="New Chat">
            +
          </button>
        </div>
        <div className="session-list">
          {sessions.map((session) => (
            <div
              key={session.id}
              className={`session-item ${session.id === currentSessionId ? 'active' : ''}`}
              onClick={() => loadSession(session.id)}
              title={formatFullDate(session.createdAt)}
            >
              <div className="session-info">
                <span className="session-title">{session.title}</span>
                <span className="session-date">{formatDate(session.updatedAt)}</span>
              </div>
              <div className="session-meta">
                <span className="message-count">{session.messageCount} msgs</span>
                <button
                  className="delete-session-btn"
                  onClick={(e) => deleteSession(session.id, e)}
                  title="Delete"
                >
                  x
                </button>
              </div>
            </div>
          ))}
          {sessions.length === 0 && (
            <div className="no-sessions">No chat history</div>
          )}
        </div>
      </div>

      {/* Toggle sidebar button */}
      <button
        className="sidebar-toggle"
        onClick={() => setShowSidebar(!showSidebar)}
      >
        {showSidebar ? '<' : '>'}
      </button>

      {/* Main chat area */}
      <div className="chat-main">
        <div className="chat-header-main">
          <div>
            <h3>{currentSession?.title || 'New Chat'}</h3>
            {currentSession && (
              <p className="chat-date">
                Started {formatFullDate(currentSession.createdAt)}
              </p>
            )}
          </div>
          {currentSession && messages.length > 0 && (
            <div className="chat-actions" style={{ position: 'relative' }}>
              <button 
                className="export-btn"
                onClick={() => setShowExportMenu(!showExportMenu)}
              >
                Export
              </button>
              {showExportMenu && (
                <div className="chat-export-menu">
                  <button onClick={() => exportSession('markdown')}>
                    Export as Markdown
                  </button>
                  <button onClick={() => exportSession('json')}>
                    Export as JSON
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Prompt Templates */}
        {messages.length === 0 && promptTemplates.length > 0 && (
          <div className="prompt-templates">
            {Object.entries(templatesByCategory).map(([category, templates]) => (
              templates.map((template) => (
                <button
                  key={template.id}
                  className="prompt-template-btn"
                  onClick={() => sendMessage(template.prompt)}
                  title={template.description}
                >
                  {template.name}
                </button>
              ))
            ))}
          </div>
        )}

        {messages.length === 0 && (
          <div className="welcome-section">
            {!isConfigured ? (
              <>
                <h2>AI Not Configured</h2>
                <p>Please configure your AI provider in Settings to use the chat feature.</p>
                <button 
                  className="configure-ai-btn"
                  onClick={onNavigateToSettings}
                >
                  Go to Settings
                </button>
              </>
            ) : (
              <>
                <h2>How can I help you today?</h2>
                <p>Ask questions about your system activity, security, and performance</p>
              </>
            )}
          </div>
        )}

        <div className="chat-messages">
          {messages.map((msg) => (
            <div
              key={msg.id}
              className={`message ${
                msg.role === 'user' ? 'message-user' :
                msg.role === 'error' ? 'message-error' :
                'message-assistant'
              }`}
            >
              <div className="message-content">{msg.content}</div>
              <div className="message-meta">
                <span className="message-time">
                  {formatDate(msg.timestamp)}
                </span>
                {msg.riskLevel && msg.riskLevel !== 'low' && (
                  <span className={`risk-badge risk-${msg.riskLevel}`}>
                    {msg.riskLevel} risk
                  </span>
                )}
              </div>
            </div>
          ))}
          {streamingMessage && (
            <div className="message message-assistant streaming">
              <div className="message-content">{streamingMessage}</div>
              <div className="streaming-indicator">
                <span className="typing-dot"></span>
                <span className="typing-dot"></span>
                <span className="typing-dot"></span>
              </div>
            </div>
          )}
          {isLoading && !streamingMessage && (
            <div className="message message-assistant">
              <div className="loading">
                <div className="spinner"></div>
                <span>Analyzing system...</span>
              </div>
            </div>
          )}
          <div ref={messagesEndRef} />
        </div>

        <div className="chat-input-container">
          {!isConfigured ? (
            <div className="chat-not-configured">
              <span>AI provider not configured. </span>
              <button onClick={onNavigateToSettings}>Configure in Settings</button>
            </div>
          ) : (
            <>
              <div className="chat-options">
                <label className="streaming-toggle">
                  <input
                    type="checkbox"
                    checked={useStreaming}
                    onChange={(e) => setUseStreaming(e.target.checked)}
                  />
                  <span>Stream responses</span>
                </label>
              </div>
              <div className="chat-input-wrapper">
                <input
                  type="text"
                  className="chat-input"
                  placeholder="Ask about your system..."
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  onKeyPress={handleKeyPress}
                  disabled={isLoading}
                />
                <button
                  className="send-btn"
                  onClick={() => sendMessage()}
                  disabled={!input.trim() || isLoading}
                >
                  Send
                </button>
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  );
}

export default Chat;
