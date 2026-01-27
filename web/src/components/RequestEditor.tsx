import React, { useState, useEffect } from 'react';
import { RequestItem, HttpMethod, Header, BodyType, FormDataItem, SidebarItem, AuthConfig } from '../types';
import AuthSelector from './AuthSelector';
import AuthConfigModal from './AuthConfigModal';

interface RequestEditorProps {
  request: RequestItem;
  onUpdate: (updates: Partial<RequestItem>) => void;
  onSend: () => void;
  isLoading: boolean;
  items?: SidebarItem[]; // For checking parent existence
}

const RequestEditor: React.FC<RequestEditorProps> = ({ request, onUpdate, onSend, isLoading, items = [] }) => {
  const [activeTab, setActiveTab] = useState<'headers' | 'body'>('headers');
  const [showAuthModal, setShowAuthModal] = useState(false);

  // Close modal if auth type changes to one that doesn't need config
  useEffect(() => {
    const auth = request.auth || { type: 'none', config: {} };
    const needsConfig = auth.type !== 'none' && auth.type !== 'inherit' && auth.type !== 'netrc';
    if (!needsConfig && showAuthModal) {
      setShowAuthModal(false);
    }
  }, [request.auth, showAuthModal]);

  const methods: HttpMethod[] = ['GET', 'POST', 'PUT', 'PATCH', 'DELETE', 'OPTIONS', 'HEAD'];
  const bodyTypes: BodyType[] = ['none', 'json', 'form-data', 'url-encoded'];

  // Normalize request data to ensure all required fields are present
  const normalizedRequest: RequestItem = {
    ...request,
    headers: request.headers || [],
    formData: request.formData || [],
    body: request.body || '',
    bodyType: request.bodyType || 'none',
    method: request.method || 'GET',
    url: request.url || '',
    auth: request.auth || { type: 'none', config: {} },
  };

  // Check if request has a parent folder
  const hasParent = normalizedRequest.parentId != null && items.some(item => item.id === normalizedRequest.parentId);

  const handleAuthSelect = (auth: AuthConfig) => {
    // If auth type requires configuration, show modal
    // Otherwise, update directly without showing modal
    const needsConfig = auth.type !== 'none' && auth.type !== 'inherit' && auth.type !== 'netrc';
    if (needsConfig) {
      // Set the auth type first, then show modal for configuration
      onUpdate({ auth });
      setShowAuthModal(true);
    } else {
      // For types that don't need config, update directly
      onUpdate({ auth });
      setShowAuthModal(false); // Ensure modal is closed
    }
  };

  const handleAuthSave = (auth: AuthConfig) => {
    onUpdate({ auth });
    setShowAuthModal(false);
  };

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

  return (
    <div className="flex flex-col h-full bg-[#131314]">
      <div className="p-6 border-b border-[#3c4043] space-y-4">
        <div className="flex items-center gap-2">
          <select
            className="bg-[#1e1e20] border border-[#3c4043] text-[#e8eaed] text-[13px] font-bold rounded-lg h-10 px-3 outline-none focus:border-[#8ab4f8] cursor-pointer"
            value={normalizedRequest.method}
            onChange={(e) => onUpdate({ method: e.target.value as HttpMethod })}
          >
            {methods.map(m => <option key={m} value={m}>{m}</option>)}
          </select>
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
        {activeTab === 'headers' ? (
          <div className="space-y-6">
            {/* Authentication Section */}
            <div className="space-y-2">
              <div className="text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                Authentication
              </div>
              <AuthSelector
                auth={normalizedRequest.auth}
                onSelect={handleAuthSelect}
                hasParent={hasParent}
              />
              {normalizedRequest.auth && 
               normalizedRequest.auth.type !== 'none' && 
               normalizedRequest.auth.type !== 'inherit' && 
               normalizedRequest.auth.type !== 'netrc' && (
                <button
                  onClick={() => {
                    // Only show modal for types that need configuration
                    const needsConfig = normalizedRequest.auth.type !== 'none' && 
                                       normalizedRequest.auth.type !== 'inherit' && 
                                       normalizedRequest.auth.type !== 'netrc';
                    if (needsConfig) {
                      setShowAuthModal(true);
                    }
                  }}
                  className="mt-2 text-[12px] font-bold text-[#8ab4f8] hover:text-[#aecbfa] flex items-center gap-1 uppercase tracking-widest transition-colors"
                >
                  <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" />
                  </svg>
                  Configure
                </button>
              )}
            </div>

            {/* Headers Section */}
            <div className="space-y-2">
              <div className="text-[11px] text-[#5f6368] uppercase font-bold tracking-widest mb-2">
                Headers
              </div>
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

      {showAuthModal && normalizedRequest.auth && (
        <AuthConfigModal
          auth={normalizedRequest.auth}
          onSave={handleAuthSave}
          onCancel={() => setShowAuthModal(false)}
        />
      )}
    </div>
  );
};

export default RequestEditor;
