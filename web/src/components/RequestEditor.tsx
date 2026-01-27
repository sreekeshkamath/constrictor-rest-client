import React, { useState, useEffect, useRef } from 'react';
import { RequestItem, HttpMethod, Header, BodyType, FormDataItem, AuthType, AuthConfig } from '../types';

interface RequestEditorProps {
  request: RequestItem;
  onUpdate: (updates: Partial<RequestItem>) => void;
  onSend: () => void;
  isLoading: boolean;
}

const RequestEditor: React.FC<RequestEditorProps> = ({ request, onUpdate, onSend, isLoading }) => {
  const [activeTab, setActiveTab] = useState<'auth' | 'headers' | 'body'>('auth');
  const [isMethodDropdownOpen, setIsMethodDropdownOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const methods: HttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'HEAD'];
  const bodyTypes: BodyType[] = ['none', 'json', 'form-data', 'url-encoded'];

  // Method-specific colors for better visibility
  const getMethodColor = (method: HttpMethod): string => {
    const colors: Record<HttpMethod, string> = {
      GET: 'text-[#81c995]',
      POST: 'text-[#fdd663]',
      PUT: 'text-[#8ab4f8]',
      PATCH: 'text-[#c58af9]',
      DELETE: 'text-[#f28b82]',
      OPTIONS: 'text-[#9aa0a6]',
      HEAD: 'text-[#9aa0a6]'
    };
    return colors[method];
  };

  const getMethodColorValue = (method: HttpMethod): string => {
    const colors: Record<HttpMethod, string> = {
      GET: '#81c995',
      POST: '#fdd663',
      PUT: '#8ab4f8',
      PATCH: '#c58af9',
      DELETE: '#f28b82',
      OPTIONS: '#9aa0a6',
      HEAD: '#9aa0a6'
    };
    return colors[method];
  };

  // Normalize request data to ensure all required fields are present
  const normalizedRequest: RequestItem = {
    ...request,
    headers: request.headers || [],
    formData: request.formData || [],
    body: request.body || '',
    bodyType: request.bodyType || 'none',
    method: request.method || 'GET',
    url: request.url || '',
    auth: request.auth || { type: 'none' },
  };

  // Sync auth config to headers when auth type changes
  useEffect(() => {
    const auth = normalizedRequest.auth || { type: 'none' };
    
    // Skip if auth is none or incomplete
    if (auth.type === 'none') {
      // Remove any existing auth headers
      const otherHeaders = normalizedRequest.headers.filter(
        h => h.key.toLowerCase() !== 'authorization'
      );
      if (otherHeaders.length !== normalizedRequest.headers.length) {
        onUpdate({ headers: otherHeaders });
      }
      return;
    }

    let authHeader: Header | null = null;
    let headerKeyToRemove = '';

    if (auth.type === 'bearer' && auth.bearerToken) {
      authHeader = { key: 'Authorization', value: `Bearer ${auth.bearerToken}`, enabled: true };
      headerKeyToRemove = 'authorization';
    } else if (auth.type === 'basic' && auth.basicUsername && auth.basicPassword) {
      const credentials = btoa(`${auth.basicUsername}:${auth.basicPassword}`);
      authHeader = { key: 'Authorization', value: `Basic ${credentials}`, enabled: true };
      headerKeyToRemove = 'authorization';
    } else if (auth.type === 'apikey' && auth.apiKeyKey && auth.apiKeyValue && auth.apiKeyLocation === 'header') {
      authHeader = { key: auth.apiKeyKey, value: auth.apiKeyValue, enabled: true };
      headerKeyToRemove = auth.apiKeyKey;
    }

    // Only update headers if we have a header to add (skip query params)
    if (authHeader) {
      // Remove old auth headers
      const otherHeaders = normalizedRequest.headers.filter(
        h => h.key.toLowerCase() !== headerKeyToRemove.toLowerCase()
      );
      
      // Check if we need to update
      const existingAuthHeader = normalizedRequest.headers.find(
        h => h.key.toLowerCase() === headerKeyToRemove.toLowerCase()
      );
      
      if (!existingAuthHeader || existingAuthHeader.value !== authHeader.value) {
        onUpdate({ headers: [...otherHeaders, authHeader] });
      }
    } else if (auth.type === 'apikey' && auth.apiKeyLocation === 'query') {
      // For query params, just remove any existing header with this key
      const otherHeaders = normalizedRequest.headers.filter(
        h => h.key !== auth.apiKeyKey
      );
      if (otherHeaders.length !== normalizedRequest.headers.length) {
        onUpdate({ headers: otherHeaders });
      }
    }
  }, [
    normalizedRequest.auth?.type,
    normalizedRequest.auth?.bearerToken,
    normalizedRequest.auth?.basicUsername,
    normalizedRequest.auth?.basicPassword,
    normalizedRequest.auth?.apiKeyKey,
    normalizedRequest.auth?.apiKeyValue,
    normalizedRequest.auth?.apiKeyLocation
  ]);

  const addHeader = () => {
    const newHeaders = [...normalizedRequest.headers, { key: '', value: '', enabled: true }];
    onUpdate({ headers: newHeaders });
  };

  const updateHeader = (index: number, field: keyof Header, value: string | boolean) => {
    const newHeaders = normalizedRequest.headers.map((h, i) => i === index ? { ...h, [field]: value } : h);
    onUpdate({ headers: newHeaders });
  };

  const removeHeader = (index: number) => {
    const newHeaders = normalizedRequest.headers.filter((_, i) => i !== index);
    onUpdate({ headers: newHeaders });
  };

  const addFormDataItem = () => {
    const newItems = [...normalizedRequest.formData, { key: '', value: '', enabled: true }];
    onUpdate({ formData: newItems });
  };

  const updateFormDataItem = (index: number, field: keyof FormDataItem, value: string | boolean) => {
    const newItems = normalizedRequest.formData.map((h, i) => i === index ? { ...h, [field]: value } : h);
    onUpdate({ formData: newItems });
  };

  const removeFormDataItem = (index: number) => {
    const newItems = normalizedRequest.formData.filter((_, i) => i !== index);
    onUpdate({ formData: newItems });
  };

  // Close dropdown on Escape key or click outside
  useEffect(() => {
    const handleEscape = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && isMethodDropdownOpen) {
        setIsMethodDropdownOpen(false);
      }
    };

    const handleClickOutside = (e: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(e.target as Node)) {
        setIsMethodDropdownOpen(false);
      }
    };

    if (isMethodDropdownOpen) {
      document.addEventListener('keydown', handleEscape);
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('keydown', handleEscape);
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isMethodDropdownOpen]);

  return (
    <div className="flex flex-col h-full bg-[#131314]">
      <div className="p-6 border-b border-[#3c4043] space-y-4">
        <div className="flex items-center gap-2">
          <div className="relative" ref={dropdownRef}>
            <button
              type="button"
              onClick={() => setIsMethodDropdownOpen(!isMethodDropdownOpen)}
              style={{ 
                backgroundColor: '#1e1e20',
                color: getMethodColorValue(normalizedRequest.method)
              }}
              className={`bg-[#1e1e20] border border-[#3c4043] ${getMethodColor(normalizedRequest.method)} text-[13px] font-bold rounded-lg h-10 px-3 pr-8 outline-none focus:border-[#8ab4f8] cursor-pointer flex items-center justify-between min-w-[100px] hover:border-[#8ab4f8] transition-colors`}
            >
              <span>{normalizedRequest.method}</span>
              <svg 
                className={`w-4 h-4 transition-transform ${isMethodDropdownOpen ? 'rotate-180' : ''}`}
                fill="none" 
                stroke="currentColor" 
                viewBox="0 0 24 24"
              >
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
              </svg>
            </button>
            {isMethodDropdownOpen && (
              <div className="absolute top-full left-0 mt-1 bg-[#1e1e20] border border-[#3c4043] rounded-lg shadow-lg z-20 min-w-[100px] overflow-hidden" style={{ backgroundColor: '#1e1e20' }}>
                {methods.map(m => (
                  <button
                    key={m}
                    type="button"
                    onClick={() => {
                      onUpdate({ method: m });
                      setIsMethodDropdownOpen(false);
                    }}
                    style={{ 
                      color: getMethodColorValue(m),
                      backgroundColor: normalizedRequest.method === m ? '#3c4043' : 'transparent'
                    }}
                    className={`w-full text-left px-3 py-2 text-[13px] font-bold ${getMethodColor(m)} hover:bg-[#3c4043] transition-colors ${
                      normalizedRequest.method === m ? 'bg-[#3c4043]' : ''
                    }`}
                  >
                    {m}
                  </button>
                ))}
              </div>
            )}
          </div>
          <input
            type="text"
            className="flex-1 bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[15px] rounded-lg h-10 px-4 outline-none focus:border-[#8ab4f8] transition-colors placeholder-[#5f6368]"
            placeholder="https://api.example.com/endpoint"
            value={normalizedRequest.url}
            onChange={(e) => onUpdate({ url: e.target.value })}
            onKeyDown={(e) => e.key === 'Enter' && onSend()}
          />
          <button
            onClick={onSend}
            disabled={isLoading}
            className="bg-[#8ab4f8] hover:bg-[#aecbfa] disabled:bg-[#3c4043] text-[#131314] font-bold text-[13px] h-10 px-6 rounded-lg transition-all active:scale-95 flex items-center gap-2 uppercase tracking-widest shadow-lg shadow-[#8ab4f8]/10"
          >
            {isLoading ? (
              <svg className="animate-spin h-4 w-4" viewBox="0 0 24 24">
                <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4" fill="none" />
                <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z" />
              </svg>
            ) : null}
            SEND
          </button>
        </div>
        <div className="text-[12px] font-bold text-[#9aa0a6] px-1 uppercase tracking-widest flex items-center gap-2">
          <span className="w-2 h-2 rounded-full bg-[#8ab4f8]"></span>
          {normalizedRequest.name}
        </div>
      </div>

      <div className="flex border-b border-[#3c4043] px-6 gap-6">
        <button
          onClick={() => setActiveTab('auth')}
          className={`py-4 text-[12px] font-bold uppercase tracking-widest border-b-2 transition-all ${
            activeTab === 'auth' ? 'border-[#8ab4f8] text-[#8ab4f8]' : 'border-transparent text-[#9aa0a6] hover:text-[#e8eaed]'
          }`}
        >
          Auth
        </button>
        <button
          onClick={() => setActiveTab('headers')}
          className={`py-4 text-[12px] font-bold uppercase tracking-widest border-b-2 transition-all ${
            activeTab === 'headers' ? 'border-[#8ab4f8] text-[#8ab4f8]' : 'border-transparent text-[#9aa0a6] hover:text-[#e8eaed]'
          }`}
        >
          Headers ({normalizedRequest.headers.length})
        </button>
        <button
          onClick={() => setActiveTab('body')}
          className={`py-4 text-[12px] font-bold uppercase tracking-widest border-b-2 transition-all ${
            activeTab === 'body' ? 'border-[#8ab4f8] text-[#8ab4f8]' : 'border-transparent text-[#9aa0a6] hover:text-[#e8eaed]'
          }`}
        >
          Body
        </button>
      </div>

      <div className="flex-1 overflow-y-auto p-6">
        {activeTab === 'auth' ? (
          <div className="space-y-6">
            <div className="flex gap-4 p-1 bg-[#1e1e20] rounded-lg w-fit">
              {(['none', 'bearer', 'basic', 'apikey'] as AuthType[]).map(type => (
                <button
                  key={type}
                  onClick={() => onUpdate({ auth: { ...normalizedRequest.auth, type } })}
                  className={`px-3 py-1.5 text-[11px] font-bold uppercase tracking-widest rounded-md transition-all ${
                    normalizedRequest.auth?.type === type
                      ? 'bg-[#8ab4f8] text-[#131314]'
                      : 'text-[#9aa0a6] hover:text-[#e8eaed]'
                  }`}
                >
                  {type === 'apikey' ? 'API Key' : type.charAt(0).toUpperCase() + type.slice(1)}
                </button>
              ))}
            </div>

            {normalizedRequest.auth?.type === 'none' && (
              <div className="flex items-center justify-center opacity-20 py-12">
                <p className="text-[12px] font-bold uppercase tracking-[0.2em]">No Authentication</p>
              </div>
            )}

            {normalizedRequest.auth?.type === 'bearer' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                    Bearer Token
                  </label>
                  <input
                    type="password"
                    className="w-full bg-[#1e1e20] border border-[#3c4043] rounded-lg px-4 py-3 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                    placeholder="Enter bearer token"
                    value={normalizedRequest.auth?.bearerToken || ''}
                    onChange={(e) => onUpdate({ auth: { ...normalizedRequest.auth, bearerToken: e.target.value } })}
                  />
                </div>
                <p className="text-[11px] text-[#5f6368] italic">
                  This will automatically add an Authorization header with Bearer token
                </p>
              </div>
            )}

            {normalizedRequest.auth?.type === 'basic' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                    Username
                  </label>
                  <input
                    type="text"
                    className="w-full bg-[#1e1e20] border border-[#3c4043] rounded-lg px-4 py-3 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                    placeholder="Enter username"
                    value={normalizedRequest.auth?.basicUsername || ''}
                    onChange={(e) => onUpdate({ auth: { ...normalizedRequest.auth, basicUsername: e.target.value } })}
                  />
                </div>
                <div>
                  <label className="block text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                    Password
                  </label>
                  <input
                    type="password"
                    className="w-full bg-[#1e1e20] border border-[#3c4043] rounded-lg px-4 py-3 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                    placeholder="Enter password"
                    value={normalizedRequest.auth?.basicPassword || ''}
                    onChange={(e) => onUpdate({ auth: { ...normalizedRequest.auth, basicPassword: e.target.value } })}
                  />
                </div>
                <p className="text-[11px] text-[#5f6368] italic">
                  This will automatically add an Authorization header with Basic authentication
                </p>
              </div>
            )}

            {normalizedRequest.auth?.type === 'apikey' && (
              <div className="space-y-4">
                <div>
                  <label className="block text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                    Key Name
                  </label>
                  <input
                    type="text"
                    className="w-full bg-[#1e1e20] border border-[#3c4043] rounded-lg px-4 py-3 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                    placeholder="e.g., X-API-Key, Api-Key"
                    value={normalizedRequest.auth?.apiKeyKey || ''}
                    onChange={(e) => onUpdate({ auth: { ...normalizedRequest.auth, apiKeyKey: e.target.value } })}
                  />
                </div>
                <div>
                  <label className="block text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                    API Key Value
                  </label>
                  <input
                    type="password"
                    className="w-full bg-[#1e1e20] border border-[#3c4043] rounded-lg px-4 py-3 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                    placeholder="Enter API key"
                    value={normalizedRequest.auth?.apiKeyValue || ''}
                    onChange={(e) => onUpdate({ auth: { ...normalizedRequest.auth, apiKeyValue: e.target.value } })}
                  />
                </div>
                <div>
                  <label className="block text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                    Location
                  </label>
                  <div className="flex gap-4 p-1 bg-[#1e1e20] rounded-lg w-fit">
                    {(['header', 'query'] as const).map(location => (
                      <button
                        key={location}
                        onClick={() => onUpdate({ auth: { ...normalizedRequest.auth, apiKeyLocation: location } })}
                        className={`px-3 py-1.5 text-[11px] font-bold uppercase tracking-widest rounded-md transition-all ${
                          (normalizedRequest.auth?.apiKeyLocation || 'header') === location
                            ? 'bg-[#8ab4f8] text-[#131314]'
                            : 'text-[#9aa0a6] hover:text-[#e8eaed]'
                        }`}
                      >
                        {location.charAt(0).toUpperCase() + location.slice(1)}
                      </button>
                    ))}
                  </div>
                </div>
                <p className="text-[11px] text-[#5f6368] italic">
                  {normalizedRequest.auth?.apiKeyLocation === 'query' 
                    ? 'This will add the API key as a query parameter in the URL'
                    : 'This will automatically add the API key as a header'}
                </p>
              </div>
            )}
          </div>
        ) : activeTab === 'headers' ? (
          <div className="space-y-2">
            <div className="grid grid-cols-[30px_1fr_1fr_40px] gap-3 mb-2 px-1">
              <div />
              <div className="text-[11px] text-[#5f6368] uppercase font-bold tracking-widest">Key</div>
              <div className="text-[11px] text-[#5f6368] uppercase font-bold tracking-widest">Value</div>
              <div />
            </div>
            {normalizedRequest.headers.map((h, i) => (
              <div key={i} className="grid grid-cols-[30px_1fr_1fr_40px] gap-3 group">
                <div className="flex items-center justify-center">
                  <input
                    type="checkbox"
                    className="accent-[#8ab4f8] rounded w-4 h-4"
                    checked={h.enabled}
                    onChange={(e) => updateHeader(i, 'enabled', e.target.checked)}
                  />
                </div>
                <input
                  className="bg-[#1e1e20] border border-[#3c4043] rounded-lg px-3 py-2 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                  placeholder="Key"
                  value={h.key}
                  onChange={(e) => updateHeader(i, 'key', e.target.value)}
                />
                <input
                  className="bg-[#1e1e20] border border-[#3c4043] rounded-lg px-3 py-2 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                  placeholder="Value"
                  value={h.value}
                  onChange={(e) => updateHeader(i, 'value', e.target.value)}
                />
                <button
                  onClick={() => removeHeader(i)}
                  className="opacity-0 group-hover:opacity-100 flex items-center justify-center text-[#5f6368] hover:text-[#f28b82] transition-opacity"
                >
                  <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                  </svg>
                </button>
              </div>
            ))}
            <button
              onClick={addHeader}
              className="mt-4 text-[12px] font-bold text-[#8ab4f8] hover:text-[#aecbfa] flex items-center gap-1 uppercase tracking-widest transition-colors"
            >
              <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeWidth="2" d="M12 4v16m8-8H4" />
              </svg>
              Add Row
            </button>
          </div>
        ) : (
          <div className="h-full flex flex-col space-y-4">
            <div className="flex gap-4 p-1 bg-[#1e1e20] rounded-lg w-fit">
              {bodyTypes.map(type => (
                <button
                  key={type}
                  onClick={() => onUpdate({ bodyType: type })}
                  className={`px-3 py-1.5 text-[11px] font-bold uppercase tracking-widest rounded-md transition-all ${
                    normalizedRequest.bodyType === type
                      ? 'bg-[#8ab4f8] text-[#131314]'
                      : 'text-[#9aa0a6] hover:text-[#e8eaed]'
                  }`}
                >
                  {type.replace('-', ' ')}
                </button>
              ))}
            </div>

            {normalizedRequest.bodyType === 'none' && (
              <div className="flex-1 flex items-center justify-center opacity-20">
                <p className="text-[12px] font-bold uppercase tracking-[0.2em]">No Request Body</p>
              </div>
            )}

            {normalizedRequest.bodyType === 'json' && (
              <textarea
                className="flex-1 bg-[#1e1e20] border border-[#3c4043] rounded-xl p-5 mono text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed] resize-none leading-relaxed shadow-inner"
                placeholder='{ "message": "hello world" }'
                value={normalizedRequest.body}
                onChange={(e) => onUpdate({ body: e.target.value })}
              />
            )}

            {(normalizedRequest.bodyType === 'form-data' || normalizedRequest.bodyType === 'url-encoded') && (
              <div className="space-y-2">
                <div className="grid grid-cols-[30px_1fr_1fr_40px] gap-3 mb-2 px-1">
                  <div />
                  <div className="text-[11px] text-[#5f6368] uppercase font-bold tracking-widest">Field Name</div>
                  <div className="text-[11px] text-[#5f6368] uppercase font-bold tracking-widest">Value</div>
                  <div />
                </div>
                {normalizedRequest.formData.map((item, i) => (
                  <div key={i} className="grid grid-cols-[30px_1fr_1fr_40px] gap-3 group">
                    <div className="flex items-center justify-center">
                      <input
                        type="checkbox"
                        className="accent-[#8ab4f8] rounded w-4 h-4"
                        checked={item.enabled}
                        onChange={(e) => updateFormDataItem(i, 'enabled', e.target.checked)}
                      />
                    </div>
                    <input
                      className="bg-[#1e1e20] border border-[#3c4043] rounded-lg px-3 py-2 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                      placeholder="Name"
                      value={item.key}
                      onChange={(e) => updateFormDataItem(i, 'key', e.target.value)}
                    />
                    <input
                      className="bg-[#1e1e20] border border-[#3c4043] rounded-lg px-3 py-2 text-[14px] outline-none focus:border-[#8ab4f8] text-[#e8eaed]"
                      placeholder="Value"
                      value={item.value}
                      onChange={(e) => updateFormDataItem(i, 'value', e.target.value)}
                    />
                    <button
                      onClick={() => removeFormDataItem(i)}
                      className="opacity-0 group-hover:opacity-100 flex items-center justify-center text-[#5f6368] hover:text-[#f28b82] transition-opacity"
                    >
                      <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                        <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" />
                      </svg>
                    </button>
                  </div>
                ))}
                <button
                  onClick={addFormDataItem}
                  className="mt-4 text-[12px] font-bold text-[#8ab4f8] hover:text-[#aecbfa] flex items-center gap-1 uppercase tracking-widest transition-colors"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeWidth="2" d="M12 4v16m8-8H4" />
                  </svg>
                  Add Parameter
                </button>
              </div>
            )}

            <div className="mt-4 text-[11px] text-[#5f6368] uppercase font-bold tracking-widest text-center">
              Body Modality: {normalizedRequest.bodyType.replace('-', ' ')}
            </div>
          </div>
        )}
      </div>
    </div>
  );
};

export default RequestEditor;
