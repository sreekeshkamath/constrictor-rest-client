import React, { useState, useEffect } from 'react';
import { AuthConfig } from '../types';

interface AuthConfigModalProps {
  auth: AuthConfig;
  onSave: (auth: AuthConfig) => void;
  onCancel: () => void;
}

const AuthConfigModal: React.FC<AuthConfigModalProps> = ({ auth, onSave, onCancel }) => {
  const [config, setConfig] = useState<Record<string, any>>(auth.config || {});

  useEffect(() => {
    setConfig(auth.config || {});
  }, [auth]);

  const handleSave = () => {
    onSave({
      ...auth,
      config,
    });
  };

  const updateConfig = (key: string, value: any) => {
    setConfig(prev => ({ ...prev, [key]: value }));
  };

  const renderConfigForm = () => {
    switch (auth.type) {
      case 'bearer':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Token
              </label>
              <input
                type="text"
                value={config.token || ''}
                onChange={(e) => updateConfig('token', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter bearer token"
              />
            </div>
          </div>
        );

      case 'basic':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Username
              </label>
              <input
                type="text"
                value={config.username || ''}
                onChange={(e) => updateConfig('username', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter username"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Password
              </label>
              <input
                type="password"
                value={config.password || ''}
                onChange={(e) => updateConfig('password', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter password"
              />
            </div>
          </div>
        );

      case 'apikey':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Key
              </label>
              <input
                type="text"
                value={config.key || ''}
                onChange={(e) => updateConfig('key', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="e.g., X-API-Key"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Value
              </label>
              <input
                type="text"
                value={config.value || ''}
                onChange={(e) => updateConfig('value', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter API key value"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Location
              </label>
              <select
                value={config.location || 'header'}
                onChange={(e) => updateConfig('location', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
              >
                <option value="header">Header</option>
                <option value="query">Query Parameter</option>
              </select>
            </div>
          </div>
        );

      case 'oauth2':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Access Token
              </label>
              <input
                type="text"
                value={config.accessToken || ''}
                onChange={(e) => updateConfig('accessToken', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter OAuth 2.0 access token"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Token Type
              </label>
              <input
                type="text"
                value={config.tokenType || 'Bearer'}
                onChange={(e) => updateConfig('tokenType', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Bearer"
              />
            </div>
          </div>
        );

      case 'oauth1':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Consumer Key
              </label>
              <input
                type="text"
                value={config.consumerKey || ''}
                onChange={(e) => updateConfig('consumerKey', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter consumer key"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Consumer Secret
              </label>
              <input
                type="password"
                value={config.consumerSecret || ''}
                onChange={(e) => updateConfig('consumerSecret', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter consumer secret"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Token
              </label>
              <input
                type="text"
                value={config.token || ''}
                onChange={(e) => updateConfig('token', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter token (optional)"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Token Secret
              </label>
              <input
                type="password"
                value={config.tokenSecret || ''}
                onChange={(e) => updateConfig('tokenSecret', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter token secret (optional)"
              />
            </div>
          </div>
        );

      case 'digest':
      case 'ntlm':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Username
              </label>
              <input
                type="text"
                value={config.username || ''}
                onChange={(e) => updateConfig('username', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter username"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Password
              </label>
              <input
                type="password"
                value={config.password || ''}
                onChange={(e) => updateConfig('password', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter password"
              />
            </div>
          </div>
        );

      case 'aws':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Access Key ID
              </label>
              <input
                type="text"
                value={config.accessKeyId || ''}
                onChange={(e) => updateConfig('accessKeyId', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter AWS access key ID"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Secret Access Key
              </label>
              <input
                type="password"
                value={config.secretAccessKey || ''}
                onChange={(e) => updateConfig('secretAccessKey', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter AWS secret access key"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Region
              </label>
              <input
                type="text"
                value={config.region || 'us-east-1'}
                onChange={(e) => updateConfig('region', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="us-east-1"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Service
              </label>
              <input
                type="text"
                value={config.service || 'execute-api'}
                onChange={(e) => updateConfig('service', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="execute-api"
              />
            </div>
          </div>
        );

      case 'hawk':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Auth ID
              </label>
              <input
                type="text"
                value={config.authId || ''}
                onChange={(e) => updateConfig('authId', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter Hawk auth ID"
              />
            </div>
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Auth Key
              </label>
              <input
                type="password"
                value={config.authKey || ''}
                onChange={(e) => updateConfig('authKey', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter Hawk auth key"
              />
            </div>
          </div>
        );

      case 'asap':
        return (
          <div className="space-y-4">
            <div>
              <label className="block text-[12px] font-bold text-[#9aa0a6] uppercase tracking-widest mb-2">
                Token
              </label>
              <input
                type="text"
                value={config.token || ''}
                onChange={(e) => updateConfig('token', e.target.value)}
                className="w-full bg-[#131314] border border-[#3c4043] text-[#e8eaed] text-[14px] rounded-lg px-3 py-2 outline-none focus:border-[#8ab4f8]"
                placeholder="Enter ASAP token"
              />
            </div>
          </div>
        );

      case 'netrc':
        return (
          <div className="space-y-4">
            <div className="text-[13px] text-[#9aa0a6]">
              Netrc authentication uses the .netrc file from your home directory. No configuration needed.
            </div>
          </div>
        );

      case 'none':
      case 'inherit':
      default:
        return (
          <div className="text-[13px] text-[#9aa0a6]">
            No configuration needed for this authentication type.
          </div>
        );
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onCancel}>
      <div className="bg-[#1e1e20] border border-[#3c4043] rounded-lg shadow-lg w-full max-w-md mx-4" onClick={(e) => e.stopPropagation()}>
        <div className="p-6 border-b border-[#3c4043]">
          <h2 className="text-[16px] font-bold text-[#e8eaed]">
            Configure {auth.type === 'bearer' ? 'Bearer Token' : auth.type === 'basic' ? 'Basic Auth' : auth.type === 'apikey' ? 'API Key' : auth.type === 'oauth2' ? 'OAuth 2.0' : auth.type === 'oauth1' ? 'OAuth 1.0' : auth.type === 'aws' ? 'AWS IAM' : auth.type === 'hawk' ? 'Hawk' : auth.type === 'asap' ? 'Atlassian ASAP' : auth.type === 'digest' ? 'Digest' : auth.type === 'ntlm' ? 'NTLM' : 'Authentication'}
          </h2>
        </div>
        <div className="p-6">
          {renderConfigForm()}
        </div>
        <div className="p-6 border-t border-[#3c4043] flex gap-3 justify-end">
          <button
            onClick={onCancel}
            className="px-4 py-2 text-[13px] font-bold text-[#9aa0a6] hover:text-[#e8eaed] bg-[#2a2b2f] rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleSave}
            className="px-4 py-2 text-[13px] font-bold text-white bg-[#8ab4f8] hover:bg-[#aecbfa] rounded-lg transition-colors"
          >
            Save
          </button>
        </div>
      </div>
    </div>
  );
};

export default AuthConfigModal;
