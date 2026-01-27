import React, { useState, useRef, useEffect } from 'react';
import { AuthConfig } from '../types';

interface AuthSelectorProps {
  auth?: AuthConfig;
  onSelect: (auth: AuthConfig) => void;
  hasParent: boolean;
}

const AUTH_TYPES = [
  { value: 'none', label: 'None', group: 'other' },
  { value: 'inherit', label: 'Inherit from parent', group: 'other' },
  { value: 'apikey', label: 'API Key', group: 'auth' },
  { value: 'basic', label: 'Basic', group: 'auth' },
  { value: 'digest', label: 'Digest', group: 'auth' },
  { value: 'ntlm', label: 'NTLM', group: 'auth' },
  { value: 'oauth1', label: 'OAuth 1.0', group: 'auth' },
  { value: 'oauth2', label: 'OAuth 2.0', group: 'auth' },
  { value: 'aws', label: 'AWS IAM', group: 'auth' },
  { value: 'bearer', label: 'Bearer Token', group: 'auth' },
  { value: 'hawk', label: 'Hawk', group: 'auth' },
  { value: 'asap', label: 'Atlassian ASAP', group: 'auth' },
  { value: 'netrc', label: 'Netrc', group: 'auth' },
] as const;

const AuthSelector: React.FC<AuthSelectorProps> = ({ auth, onSelect, hasParent }) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  const currentAuthType = auth?.type || 'none';
  const currentAuth = AUTH_TYPES.find(a => a.value === currentAuthType);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }

    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleSelect = (type: string) => {
    const newAuth: AuthConfig = {
      type: type as AuthConfig['type'],
      config: {},
    };
    onSelect(newAuth);
    setIsOpen(false);
  };

  const otherOptions = AUTH_TYPES.filter(a => a.group === 'other' && (a.value !== 'inherit' || hasParent));
  const authOptions = AUTH_TYPES.filter(a => a.group === 'auth');

  return (
    <div className="relative" ref={dropdownRef}>
      <button
        onClick={() => setIsOpen(!isOpen)}
        className="w-full bg-[#1e1e20] border border-[#3c4043] text-[#e8eaed] text-[13px] rounded-lg h-10 px-3 outline-none focus:border-[#8ab4f8] cursor-pointer flex items-center justify-between"
      >
        <span className="flex items-center gap-2">
          {currentAuth ? (
            <>
              <span>{currentAuth.label}</span>
            </>
          ) : (
            <span className="text-[#5f6368]">Select Auth Type</span>
          )}
        </span>
        <svg
          className={`w-4 h-4 text-[#9aa0a6] transition-transform ${isOpen ? 'rotate-180' : ''}`}
          fill="none"
          stroke="currentColor"
          viewBox="0 0 24 24"
        >
          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M19 9l-7 7-7-7" />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute z-50 mt-1 w-full bg-[#1e1e20] border border-[#3c4043] rounded-lg shadow-lg overflow-hidden">
          <div className="py-1">
            {/* OTHER section */}
            {otherOptions.length > 0 && (
              <>
                <div className="px-4 py-2 text-[11px] font-bold text-[#5f6368] uppercase tracking-widest border-b border-[#3c4043]">
                  OTHER
                </div>
                {otherOptions.map((option) => (
                  <button
                    key={option.value}
                    onClick={() => handleSelect(option.value)}
                    className={`w-full text-left px-4 py-2 text-[13px] text-[#e8eaed] hover:bg-[#2a2b2f] flex items-center justify-between ${
                      currentAuthType === option.value ? 'bg-[#2a2b2f]' : ''
                    }`}
                  >
                    <span>{option.label}</span>
                    {currentAuthType === option.value && (
                      <svg className="w-4 h-4 text-[#34a853]" fill="currentColor" viewBox="0 0 20 20">
                        <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                      </svg>
                    )}
                  </button>
                ))}
              </>
            )}

            {/* AUTH TYPES section */}
            <div className="px-4 py-2 text-[11px] font-bold text-[#5f6368] uppercase tracking-widest border-b border-[#3c4043] flex items-center gap-2">
              <svg className="w-3 h-3" fill="currentColor" viewBox="0 0 20 20">
                <path fillRule="evenodd" d="M5 9V7a5 5 0 0110 0v2a2 2 0 012 2v5a2 2 0 01-2 2H5a2 2 0 01-2-2v-5a2 2 0 012-2zm8-2v2H7V7a3 3 0 016 0z" clipRule="evenodd" />
              </svg>
              AUTH TYPES
            </div>
            {authOptions.map((option) => (
              <button
                key={option.value}
                onClick={() => handleSelect(option.value)}
                className={`w-full text-left px-4 py-2 text-[13px] text-[#e8eaed] hover:bg-[#2a2b2f] flex items-center justify-between ${
                  currentAuthType === option.value ? 'bg-[#2a2b2f]' : ''
                }`}
              >
                <span>{option.label}</span>
                {currentAuthType === option.value && (
                  <svg className="w-4 h-4 text-[#34a853]" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                )}
              </button>
            ))}
          </div>
        </div>
      )}
    </div>
  );
};

export default AuthSelector;
