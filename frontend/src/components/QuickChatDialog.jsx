import React, { useState, useEffect, useRef } from 'react';
import { useErrorDialog } from '../contexts/ErrorDialogContext';

export default function QuickChatDialog({ isOpen, onClose }) {
  const { showError } = useErrorDialog();
  
  const [input, setInput] = useState('');
  const [isLoading, setIsLoading] = useState(false);
  const [messages, setMessages] = useState([]);
  const [currentSessionId, setCurrentSessionId] = useState(null);
  const [sessions, setSessions] = useState([]);
  const [showSessionList, setShowSessionList] = useState(false);
  const [isClosing, setIsClosing] = useState(false);
  const messagesEndRef = useRef(null);
  const inputRef = useRef(null);

  // Load sessions on mount
  useEffect(() => {
    const loadSessions = async () => {
      try {
        if (window.go?.main?.App?.GetAllChatSessions) {
          const allSessions = await window.go.main.App.GetAllChatSessions();
          setSessions(allSessions || []);
          if (allSessions && allSessions.length > 0 && !currentSessionId) {
            setCurrentSessionId(allSessions[0].id);
            loadSessionMessages(allSessions[0].id);
          }
        }
      } catch (err) {
        console.error('Error loading sessions:', err);
      }
    };
    
    loadSessions();
  }, []);

  // Load messages for a session
  const loadSessionMessages = async (sessionId) => {
    try {
      if (window.go?.main?.App?.GetChatSession) {
        const session = await window.go.main.App.GetChatSession(sessionId);
        if (session?.messages) {
          setMessages(session.messages);
          scrollToBottom();
        }
      }
    } catch (err) {
      console.error('Error loading session:', err);
    }
  };

  // Switch to a different session
  const switchSession = async (sessionId) => {
    setCurrentSessionId(sessionId);
    await loadSessionMessages(sessionId);
    setShowSessionList(false);
  };

  // Scroll to bottom
  const scrollToBottom = () => {
    setTimeout(() => {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, 0);
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages]);

  // Focus input when dialog opens
  useEffect(() => {
    if (isOpen) {
      setTimeout(() => inputRef.current?.focus(), 100);
    }
  }, [isOpen]);

  // Handle send message
  const handleSend = async () => {
    if (!input.trim() || isLoading) return;

    const question = input.trim();
    setInput('');
    setIsLoading(true);

    try {
      let sessionId = currentSessionId;

      // Create session if needed
      if (!sessionId) {
        if (window.go?.main?.App?.CreateChatSession) {
          const session = await window.go.main.App.CreateChatSession('');
          if (session) {
            sessionId = session.id;
            setCurrentSessionId(sessionId);
            setSessions(prev => [session, ...prev]);
          }
        }
      }

      if (!sessionId) {
        showError('Error', 'Failed to create chat session');
        return;
      }

      // Add user message to UI immediately
      const userMsg = {
        id: 'user-' + Date.now(),
        role: 'user',
        content: question,
        timestamp: Date.now(),
      };
      setMessages(prev => [...prev, userMsg]);

      // Send message
      if (window.go?.main?.App?.SendChatMessage) {
        await window.go.main.App.SendChatMessage(sessionId, question);
        
        // Reload session to get updated messages
        if (window.go?.main?.App?.GetChatSession) {
          const session = await window.go.main.App.GetChatSession(sessionId);
          if (session?.messages) {
            setMessages(session.messages);
          }
        }
      }
    } catch (err) {
      console.error('Error sending message:', err);
      showError('Error', err.message || 'Failed to send message');
    } finally {
      setIsLoading(false);
    }
  };

  const handleKeyPress = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  const handleClose = () => {
    setIsClosing(true);
    setTimeout(() => {
      setIsClosing(false);
      onClose();
    }, 300);
  };

  if (!isOpen) return null;

  const currentSession = sessions.find(s => s.id === currentSessionId);

  return (
    <div className={`quick-chat-dialog ${isClosing ? 'closing' : ''}`}>
      {/* Header */}
      <div className="quick-chat-header">
        <div>
          <h3>AI Assistant</h3>
          {currentSession && (
            <p style={{ margin: 0, fontSize: '12px', color: 'var(--text-muted)' }}>
              {currentSession.title}
            </p>
          )}
        </div>
        <div style={{ display: 'flex', gap: '8px', alignItems: 'center' }}>
          <button 
            className="quick-chat-close-btn"
            onClick={() => setShowSessionList(!showSessionList)}
            title="Chat Sessions"
            style={{
              fontSize: '16px',
              padding: '4px 8px',
              background: showSessionList ? 'var(--bg-tertiary)' : 'transparent',
            }}
          >
            ☰
          </button>
          <button 
            className="quick-chat-close-btn"
            onClick={handleClose}
            title="Close (Ctrl+K)"
          >
            ✕
          </button>
        </div>
      </div>

      {/* Session List Dropdown */}
      {showSessionList && (
        <div style={{
          borderBottom: '1px solid var(--border-color)',
          maxHeight: '150px',
          overflowY: 'auto',
          padding: '8px',
          background: 'var(--bg-primary)',
        }}>
          {sessions.length === 0 ? (
            <div style={{ padding: '8px', color: 'var(--text-muted)', fontSize: '12px' }}>
              No chat sessions
            </div>
          ) : (
            sessions.map(session => (
              <button
                key={session.id}
                onClick={() => switchSession(session.id)}
                style={{
                  width: '100%',
                  padding: '8px 12px',
                  marginBottom: '4px',
                  border: '1px solid var(--border-color)',
                  borderRadius: '6px',
                  background: currentSessionId === session.id ? 'var(--accent-primary)' : 'transparent',
                  color: currentSessionId === session.id ? 'white' : 'var(--text-primary)',
                  cursor: 'pointer',
                  textAlign: 'left',
                  fontSize: '12px',
                  transition: 'all 0.2s',
                }}
                onMouseEnter={(e) => {
                  if (currentSessionId !== session.id) {
                    e.target.style.background = 'var(--bg-tertiary)';
                  }
                }}
                onMouseLeave={(e) => {
                  if (currentSessionId !== session.id) {
                    e.target.style.background = 'transparent';
                  }
                }}
              >
                {session.title}
              </button>
            ))
          )}
        </div>
      )}

      {/* Messages */}
      <div className="quick-chat-messages">
        {messages.length === 0 ? (
          <div style={{ textAlign: 'center', padding: '20px', color: 'var(--text-muted)' }}>
            <p style={{ marginBottom: '4px' }}>Start a conversation</p>
            <p style={{ fontSize: '12px', margin: 0 }}>Type your question below</p>
          </div>
        ) : (
          <>
            {messages.map((msg) => (
              <div 
                key={msg.id} 
                className={msg.role === 'user' ? 'message message-user' : 'message message-assistant'}
                style={{ maxWidth: '90%' }}
              >
                <div className="message-content">{msg.content}</div>
              </div>
            ))}
            {isLoading && (
              <div className="message message-assistant" style={{ maxWidth: '90%' }}>
                <div className="loading">
                  <div className="spinner"></div>
                  <span style={{ fontSize: '12px' }}>Analyzing system...</span>
                </div>
              </div>
            )}
            <div ref={messagesEndRef} />
          </>
        )}
      </div>

      {/* Input */}
      <div className="quick-chat-input-container">
        <div className="quick-chat-input-wrapper">
          <input
            ref={inputRef}
            type="text"
            className="quick-chat-input"
            placeholder="Ask about your system..."
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyPress={handleKeyPress}
            disabled={isLoading}
          />
          <button
            className="quick-chat-send-btn"
            onClick={handleSend}
            disabled={!input.trim() || isLoading}
            title="Send (Enter)"
          >
            ↑
          </button>
        </div>
      </div>
    </div>
  );
}
