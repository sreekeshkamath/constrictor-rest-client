import React, { useState } from 'react';
import { AppSettings } from '../types';

interface SettingsModalProps {
  settings: AppSettings;
  onUpdate: (settings: AppSettings) => void;
  onClose: () => void;
}

const SettingsModal: React.FC<SettingsModalProps> = ({ settings, onUpdate, onClose }) => {
  const [isAuthenticating, setIsAuthenticating] = useState(false);
  const [isBackingUp, setIsBackingUp] = useState(false);
  const [backupStatus, setBackupStatus] = useState<string | null>(null);

  const handleGDriveToggle = () => {
    onUpdate({
      ...settings,
      gdrive: { ...settings.gdrive, enabled: !settings.gdrive.enabled }
    });
  };

  const updateGDField = (field: 'apiKey' | 'clientId' | 'accessToken', value: string) => {
    onUpdate({
      ...settings,
      gdrive: { ...settings.gdrive, [field]: value }
    });
  };

  const handleAuthenticate = () => {
    if (!settings.gdrive.clientId) {
      alert('Please enter your OAuth Client ID first');
      return;
    }

    // Store client ID in sessionStorage for callback handling
    sessionStorage.setItem('gdrive_client_id', settings.gdrive.clientId);
    sessionStorage.setItem('gdrive_redirect', 'true');

    // Redirect to Google OAuth
    const scope = 'https://www.googleapis.com/auth/drive.file';
    const redirectUri = window.location.origin + window.location.pathname;
    const responseType = 'token';
    const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${encodeURIComponent(settings.gdrive.clientId)}&redirect_uri=${encodeURIComponent(redirectUri)}&response_type=${responseType}&scope=${encodeURIComponent(scope)}&access_type=offline&prompt=consent`;

    window.location.href = authUrl;
  };

  // Check for OAuth callback on mount
  React.useEffect(() => {
    const hash = window.location.hash;
    if (hash && sessionStorage.getItem('gdrive_redirect') === 'true') {
      const params = new URLSearchParams(hash.substring(1));
      const accessToken = params.get('access_token');
      const error = params.get('error');

      sessionStorage.removeItem('gdrive_redirect');
      window.history.replaceState(null, '', window.location.pathname);

      if (accessToken) {
        updateGDField('accessToken', accessToken);
        setBackupStatus('Authentication successful!');
      } else if (error) {
        setBackupStatus(`Authentication failed: ${error}`);
      }
    }
  }, []);

  const handleBackup = async () => {
    if (!settings.gdrive.accessToken) {
      alert('Please authenticate with Google Drive first');
      return;
    }

    setIsBackingUp(true);
    setBackupStatus('Backing up workspace...');

    try {
      const response = await fetch('/api/gdrive/backup', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          accessToken: settings.gdrive.accessToken,
          filename: `constrictor-workspace-${new Date().toISOString().split('T')[0]}.json`,
        }),
      });

      if (!response.ok) {
        const error = await response.json();
        throw new Error(error.message || 'Backup failed');
      }

      const result = await response.json();
      setBackupStatus(`Backup successful! File ID: ${result.fileId}`);

      // Update last sync time
      onUpdate({
        ...settings,
        gdrive: { ...settings.gdrive, lastSync: Date.now().toString() }
      });
    } catch (error: any) {
      console.error('Backup error:', error);
      setBackupStatus(`Backup failed: ${error.message}`);
    } finally {
      setIsBackingUp(false);
    }
  };


  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
      <div className="w-full max-w-lg bg-[#1e1e20] border border-[#3c4043] rounded-2xl shadow-2xl overflow-hidden">
        <div className="p-6 border-b border-[#3c4043] flex justify-between items-center">
          <h2 className="heading-bold text-[13px] tracking-widest text-[#e8eaed]">Settings</h2>
          <button onClick={onClose} className="text-[#9aa0a6] hover:text-[#e8eaed] transition-colors p-1">
            <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M6 18L18 6M6 6l12 12" /></svg>
          </button>
        </div>

        <div className="p-8 space-y-10">
          <section className="space-y-6">
            <div className="flex items-center justify-between">
              <div>
                <h3 className="heading-bold text-[12px] tracking-widest text-[#e8eaed]">Google Drive Sync</h3>
                <p className="text-[12px] text-[#9aa0a6] font-medium mt-1">Mirror your requests to cloud storage</p>
              </div>
              <button
                onClick={handleGDriveToggle}
                className={`w-12 h-6 rounded-full transition-colors relative ${settings.gdrive.enabled ? 'bg-[#8ab4f8]' : 'bg-[#3c4043]'}`}
              >
                <div className={`absolute top-1 left-1 w-4 h-4 rounded-full bg-[#131314] transition-transform ${settings.gdrive.enabled ? 'translate-x-6' : 'translate-x-0'}`} />
              </button>
            </div>

            {settings.gdrive.enabled && (
              <div className="space-y-5 pt-4 border-t border-[#3c4043]">
                <p className="text-[10px] text-[#5f6368] font-bold italic uppercase">* Tokens never touch our servers - OAuth handled entirely in browser</p>
                <div className="space-y-2">
                  <label className="text-[11px] font-bold uppercase tracking-widest text-[#9aa0a6]">OAuth Client ID</label>
                  <input
                    type="text"
                    className="w-full bg-[#131314] border border-[#3c4043] rounded-lg p-3 text-[14px] text-[#e8eaed] outline-none focus:border-[#8ab4f8] transition-colors"
                    placeholder="Enter Client ID..."
                    value={settings.gdrive.clientId}
                    onChange={(e) => updateGDField('clientId', e.target.value)}
                  />
                </div>
                <button
                  onClick={handleAuthenticate}
                  disabled={isAuthenticating || !settings.gdrive.clientId}
                  className="w-full py-4 bg-[#3c4043] hover:bg-[#4d5156] disabled:opacity-50 disabled:cursor-not-allowed text-[#e8eaed] text-[12px] font-bold uppercase tracking-widest transition-colors rounded-xl border border-[#5f6368]/20"
                >
                  {isAuthenticating ? 'Authenticating...' : settings.gdrive.accessToken ? 'Re-authenticate' : 'Authenticate with Google Drive'}
                </button>

                {settings.gdrive.accessToken && (
                  <button
                    onClick={handleBackup}
                    disabled={isBackingUp}
                    className="w-full py-4 bg-[#8ab4f8] hover:bg-[#aecbfa] disabled:opacity-50 disabled:cursor-not-allowed text-[#131314] text-[12px] font-bold uppercase tracking-widest transition-colors rounded-xl border border-[#5f6368]/20"
                  >
                    {isBackingUp ? 'Backing up...' : 'Backup Workspace Now'}
                  </button>
                )}

                {backupStatus && (
                  <div className={`p-3 rounded-lg text-[12px] font-medium ${
                    backupStatus.includes('successful') || backupStatus.includes('success')
                      ? 'bg-[#137333] text-[#81c995]'
                      : backupStatus.includes('failed') || backupStatus.includes('error')
                      ? 'bg-[#8e0000] text-[#f28b82]'
                      : 'bg-[#3c4043] text-[#9aa0a6]'
                  }`}>
                    {backupStatus}
                  </div>
                )}

                {settings.gdrive.lastSync && (
                  <p className="text-[10px] text-[#5f6368] font-bold uppercase">
                    Last synced: {new Date(parseInt(settings.gdrive.lastSync)).toLocaleString()}
                  </p>
                )}
              </div>
            )}
          </section>

          <section>
             <h3 className="heading-bold text-[12px] tracking-widest text-[#e8eaed] mb-4">Appearance</h3>
             <div className="p-6 bg-[#131314] rounded-xl border border-[#3c4043]">
               <span className="heading-bold text-xs tracking-widest block mb-2 text-[#8ab4f8]">Montserrat Geometric</span>
               <p className="text-[12px] text-[#9aa0a6] font-medium leading-relaxed">High-legibility geometric sans-serif for optimal API development.</p>
             </div>
          </section>
        </div>

        <div className="p-6 bg-[#131314] border-t border-[#3c4043]">
           <button onClick={onClose} className="w-full py-4 bg-[#8ab4f8] hover:bg-[#aecbfa] text-[#131314] text-[13px] font-bold uppercase tracking-widest transition-all rounded-xl shadow-lg active:scale-[0.98]">
             Apply Configurations
           </button>
        </div>
      </div>
    </div>
  );
};

export default SettingsModal;
