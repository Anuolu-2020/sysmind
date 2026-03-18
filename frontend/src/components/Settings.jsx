import { useState, useEffect, useCallback } from 'react';
import { useTheme, THEME_LIST, FONTS, FONT_SIZES } from '../contexts/ThemeContext';

function Settings({ onConfigChange }) {
  const { theme, setTheme, font, setFont, fontSize, setFontSize } = useTheme();
  const [providers, setProviders] = useState([]);
  const [config, setConfig] = useState({
    provider: 'openai',
    model: 'gpt-3.5-turbo',
    apiKey: '',
    cloudflareAcct: '',
    localEndpoint: 'http://localhost:11434',
  });
  const [isSaving, setIsSaving] = useState(false);
  const [saveStatus, setSaveStatus] = useState('');
  const [versionInfo, setVersionInfo] = useState(null);

  const loadData = useCallback(async () => {
    try {
      if (window.go?.main?.App) {
        const [provs, cfg, version] = await Promise.all([
          window.go.main.App.GetAvailableProviders(),
          window.go.main.App.GetAIConfig(),
          window.go.main.App.GetVersion().catch(() => null),
        ]);
        setProviders(provs || []);
        if (cfg) {
          setConfig(cfg);
        }
        if (version) {
          setVersionInfo(version);
        }
      }
    } catch (err) {
      console.error('Error loading settings:', err);
    }
  }, []);

  useEffect(() => {
    loadData();
  }, [loadData]);

  const handleSave = async () => {
    setIsSaving(true);
    setSaveStatus('');

    try {
      if (window.go?.main?.App?.SetAIConfig) {
        await window.go.main.App.SetAIConfig(config);
        setSaveStatus('Settings saved successfully!');
        if (onConfigChange) {
          onConfigChange();
        }
      }
    } catch (err) {
      setSaveStatus('Error saving settings: ' + err.message);
    } finally {
      setIsSaving(false);
      setTimeout(() => setSaveStatus(''), 3000);
    }
  };

  const selectedProvider = providers.find(p => p.id === config.provider);
  const models = selectedProvider?.models || [];

  const getProviderInfo = (providerId) => {
    const info = {
      openai: {
        title: 'OpenAI',
        description: 'Get your API key from platform.openai.com',
        link: 'https://platform.openai.com/api-keys',
        tip: 'GPT-4o for best results, GPT-4o-mini for faster responses.',
      },
      anthropic: {
        title: 'Anthropic (Claude)',
        description: 'Get your API key from console.anthropic.com',
        link: 'https://console.anthropic.com/settings/keys',
        tip: 'Claude 3.5 Sonnet offers excellent reasoning and analysis.',
      },
      kimi: {
        title: 'Kimi (Moonshot AI)',
        description: 'Get your API key from platform.moonshot.cn',
        link: 'https://platform.moonshot.cn/console/api-keys',
        tip: 'Excellent for Chinese and English. 128K context window.',
      },
      glm: {
        title: 'GLM (Zhipu AI)',
        description: 'Get your API key from open.bigmodel.cn',
        link: 'https://open.bigmodel.cn/usercenter/apikeys',
        tip: 'Strong Chinese language model with vision capabilities.',
      },
      copilot: {
        title: 'GitHub Copilot',
        description: 'Requires GitHub Copilot subscription. Uses your GitHub token.',
        link: 'https://github.com/settings/copilot',
        tip: 'Uses the same models as GitHub Copilot Chat.',
      },
      cloudflare: {
        title: 'Cloudflare Workers AI',
        description: 'Get your API token from the Cloudflare dashboard.',
        link: 'https://dash.cloudflare.com',
        tip: 'Free tier includes generous limits for personal use.',
      },
      local: {
        title: 'Local LLM (Ollama)',
        description: 'Run models locally with Ollama.',
        link: 'https://ollama.ai',
        tip: 'Run: ollama pull llama3.2 && ollama serve',
      },
    };
    return info[providerId] || { title: providerId, description: '', tip: '' };
  };

  const providerInfo = getProviderInfo(config.provider);

  return (
    <div className="panel settings-panel">
      <h2 style={{ marginBottom: '20px' }}>Settings</h2>

      {/* Theme Section */}
      <div className="settings-section">
        <h3>Appearance</h3>
        
        <div className="form-group">
          <label>Theme</label>
          <div className="theme-grid">
            {THEME_LIST.map((t) => (
              <button
                key={t.id}
                className={`theme-card ${theme === t.id ? 'active' : ''}`}
                onClick={() => setTheme(t.id)}
                title={t.name}
              >
                <div className={`theme-preview theme-preview-${t.id}`} />
                <span className="theme-name">{t.name}</span>
              </button>
            ))}
          </div>
        </div>

        <div className="form-row">
          <div className="form-group">
            <label>Font Family</label>
            <select
              value={font}
              onChange={(e) => setFont(e.target.value)}
            >
              {FONTS.map((f) => (
                <option key={f.id} value={f.id}>
                  {f.name}
                </option>
              ))}
            </select>
          </div>

          <div className="form-group">
            <label>Font Size</label>
            <select
              value={fontSize}
              onChange={(e) => setFontSize(e.target.value)}
            >
              {FONT_SIZES.map((s) => (
                <option key={s.id} value={s.id}>
                  {s.name} ({s.value})
                </option>
              ))}
            </select>
          </div>
        </div>

        <p className="font-preview" style={{ marginTop: '12px', padding: '12px', background: 'var(--bg-secondary)', borderRadius: '8px' }}>
          The quick brown fox jumps over the lazy dog. 0123456789
        </p>
      </div>

      {/* AI Provider Section */}
      <div className="settings-section">
        <h3>AI Provider</h3>
        
        <div className="form-group">
          <label>Provider</label>
          <select
            value={config.provider}
            onChange={(e) => {
              const newProvider = e.target.value;
              const provInfo = providers.find(p => p.id === newProvider);
              setConfig({
                ...config,
                provider: newProvider,
                model: provInfo?.models[0]?.id || '',
              });
            }}
          >
            {providers.map((prov) => (
              <option key={prov.id} value={prov.id}>
                {prov.name}
              </option>
            ))}
          </select>
        </div>

        <div className="form-group">
          <label>Model</label>
          <select
            value={config.model}
            onChange={(e) => setConfig({ ...config, model: e.target.value })}
          >
            {models.map((model) => (
              <option key={model.id} value={model.id}>
                {model.name}
              </option>
            ))}
          </select>
        </div>

        {selectedProvider?.requiresApiKey && (
          <div className="form-group">
            <label>API Key</label>
            <input
              type="password"
              value={config.apiKey}
              onChange={(e) => setConfig({ ...config, apiKey: e.target.value })}
              placeholder="Enter your API key"
            />
          </div>
        )}

        {selectedProvider?.requiresAcctId && (
          <div className="form-group">
            <label>Account ID</label>
            <input
              type="text"
              value={config.cloudflareAcct}
              onChange={(e) => setConfig({ ...config, cloudflareAcct: e.target.value })}
              placeholder="Your account ID"
            />
          </div>
        )}

        {selectedProvider?.requiresEndpoint && (
          <div className="form-group">
            <label>Endpoint URL</label>
            <input
              type="text"
              value={config.localEndpoint}
              onChange={(e) => setConfig({ ...config, localEndpoint: e.target.value })}
              placeholder="http://localhost:11434"
            />
          </div>
        )}

        <button className="save-btn" onClick={handleSave} disabled={isSaving}>
          {isSaving ? 'Saving...' : 'Save Settings'}
        </button>

        {saveStatus && (
          <div
            className={`status-indicator ${saveStatus.includes('Error') ? 'status-not-configured' : 'status-configured'}`}
            style={{ marginTop: '12px' }}
          >
            {saveStatus}
          </div>
        )}
      </div>

      <div className="settings-section">
        <h3>Provider Information</h3>
        <div className="provider-info">
          <p className="provider-title">{providerInfo.title}</p>
          <p className="provider-desc">
            {providerInfo.description}
            {providerInfo.link && (
              <>
                {' '}
                <a href={providerInfo.link} target="_blank" rel="noreferrer">
                  Get API Key
                </a>
              </>
            )}
          </p>
          {providerInfo.tip && (
            <p className="provider-tip">{providerInfo.tip}</p>
          )}
        </div>
      </div>

      <div className="settings-section">
        <h3>About SysMind</h3>
        {versionInfo && (
          <div className="version-info">
            <div className="version-row">
              <span className="version-label">Version:</span>
              <span className="version-value">{versionInfo.version}</span>
            </div>
            {versionInfo.gitTag && versionInfo.gitTag !== "unknown" && (
              <div className="version-row">
                <span className="version-label">Release:</span>
                <span className="version-value">{versionInfo.gitTag}</span>
              </div>
            )}
            <div className="version-row">
              <span className="version-label">Build:</span>
              <span className="version-value">{versionInfo.gitCommit?.substring(0, 8) || 'unknown'}</span>
            </div>
            <div className="version-row">
              <span className="version-label">Platform:</span>
              <span className="version-value">{versionInfo.platform}/{versionInfo.arch}</span>
            </div>
            <div className="version-row">
              <span className="version-label">Go Version:</span>
              <span className="version-value">{versionInfo.goVersion}</span>
            </div>
            {versionInfo.buildDate && versionInfo.buildDate !== "unknown" && (
              <div className="version-row">
                <span className="version-label">Built:</span>
                <span className="version-value">
                  {new Date(versionInfo.buildDate).toLocaleDateString()}
                </span>
              </div>
            )}
          </div>
        )}
        <div className="about-description">
          <p>AI-powered system monitoring assistant that helps you understand what your computer is doing in real-time.</p>
          <div className="about-links">
            <a href="https://github.com/yourusername/sysmind" target="_blank" rel="noreferrer">
              GitHub Repository
            </a>
            <a href="https://github.com/yourusername/sysmind/issues" target="_blank" rel="noreferrer">
              Report Issue
            </a>
            <a href="https://github.com/yourusername/sysmind/blob/main/LICENSE" target="_blank" rel="noreferrer">
              License (MIT)
            </a>
          </div>
        </div>
      </div>
    </div>
  );
}

export default Settings;
