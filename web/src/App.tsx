import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import RequestEditor from './components/RequestEditor';
import ResponseViewer from './components/ResponseViewer';
import SettingsModal from './components/SettingsModal';
import { SidebarItem, RequestItem, ResponseData, AppSettings } from './types';
import { v4 as uuidv4 } from 'uuid';

const API_BASE = '/api';

const App: React.FC = () => {
  const [items, setItems] = useState<SidebarItem[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [response, setResponse] = useState<ResponseData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [settings, setSettings] = useState<AppSettings>({
    gdrive: { enabled: false, apiKey: '', clientId: '' }
  });
  const [syncStatus, setSyncStatus] = useState<'idle' | 'syncing' | 'synced' | 'error'>('idle');

  // Load settings from localStorage
  useEffect(() => {
    const savedSettings = localStorage.getItem('constrictor_settings');
    if (savedSettings) {
      try {
        setSettings(JSON.parse(savedSettings));
      } catch (e) {
        console.error('Failed to load settings:', e);
      }
    }
  }, []);

  // Save settings to localStorage
  useEffect(() => {
    localStorage.setItem('constrictor_settings', JSON.stringify(settings));
  }, [settings]);

  // Load workspace from backend
  useEffect(() => {
    const loadWorkspace = async () => {
      try {
        const res = await fetch(`${API_BASE}/workspace`);
        if (!res.ok) throw new Error('Failed to load workspace');
        const data = await res.json();

        if (data.items && data.items.length > 0) {
          setItems(data.items);
          if (data.items[0].type === 'request') {
            setActiveId(data.items[0].id);
          }
        } else {
          // Create default request
          const defaultReq: RequestItem = {
            id: 'welcome-req',
            name: 'Get Users Demo',
            method: 'GET',
            url: 'https://jsonplaceholder.typicode.com/users',
            headers: [{ key: 'Content-Type', value: 'application/json', enabled: true }],
            bodyType: 'json',
            body: '',
            formData: [],
            type: 'request',
            createdAt: Date.now()
          };
          setItems([defaultReq]);
          setActiveId(defaultReq.id);
        }
      } catch (err: any) {
        console.error('Failed to load workspace:', err);
        // Create default request on error
        const defaultReq: RequestItem = {
          id: 'welcome-req',
          name: 'Get Users Demo',
          method: 'GET',
          url: 'https://jsonplaceholder.typicode.com/users',
          headers: [{ key: 'Content-Type', value: 'application/json', enabled: true }],
          bodyType: 'json',
          body: '',
          formData: [],
          type: 'request',
          createdAt: Date.now()
        };
        setItems([defaultReq]);
        setActiveId(defaultReq.id);
      }
    };

    loadWorkspace();
  }, []);

  // Save workspace to backend when items change, and sync to Google Drive if configured
  useEffect(() => {
    if (items.length === 0) return;

    const saveWorkspace = async () => {
      try {
        // Save to local backend
        await fetch(`${API_BASE}/workspace`, {
          method: 'PUT',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ version: 1, items })
        });

        // Auto-sync to Google Drive if configured
        if (settings.gdrive.enabled && settings.gdrive.accessToken) {
          setSyncStatus('syncing');
          try {
            const syncResponse = await fetch(`${API_BASE}/gdrive/backup`, {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({
                accessToken: settings.gdrive.accessToken,
                filename: `constrictor-workspace-${new Date().toISOString().split('T')[0]}.json`,
              }),
            });

            if (syncResponse.ok) {
              setSyncStatus('synced');
              // Update last sync time
              setSettings(prev => ({
                ...prev,
                gdrive: { ...prev.gdrive, lastSync: Date.now().toString() }
              }));
              // Reset status after 3 seconds
              setTimeout(() => setSyncStatus('idle'), 3000);
            } else {
              setSyncStatus('error');
              setTimeout(() => setSyncStatus('idle'), 3000);
            }
          } catch (syncErr) {
            console.error('Failed to sync to Google Drive:', syncErr);
            setSyncStatus('error');
            setTimeout(() => setSyncStatus('idle'), 3000);
          }
        }
      } catch (err) {
        console.error('Failed to save workspace:', err);
      }
    };

    const timeoutId = setTimeout(saveWorkspace, 500); // Debounce
    return () => clearTimeout(timeoutId);
  }, [items, settings.gdrive.enabled, settings.gdrive.accessToken]);

  const activeItem = items.find(i => i.id === activeId) as RequestItem | undefined;

  const updateActiveRequest = (updates: Partial<RequestItem>) => {
    if (!activeId) return;
    setItems(prev => prev.map(item => item.id === activeId ? { ...item, ...updates } : item));
  };

  const handleSendRequest = async () => {
    if (!activeItem) return;
    setIsLoading(true);
    setError(null);
    setResponse(null);

    try {
      const res = await fetch(`${API_BASE}/execute`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          method: activeItem.method,
          url: activeItem.url,
          headers: activeItem.headers,
          bodyType: activeItem.bodyType,
          body: activeItem.body,
          formData: activeItem.formData
        })
      });

      if (!res.ok) {
        const errorData = await res.json();
        throw new Error(errorData.message || 'Request failed');
      }

      const data = await res.json();

      if (data.error) {
        setError(data.error.message || 'Request failed');
      } else {
        setResponse({
          status: data.status,
          statusText: data.statusText,
          headers: data.headers,
          body: data.body,
          time: data.timeMs,
          size: data.sizeBytes
        });
      }
    } catch (err: any) {
      setError(err.message || "Failed to execute request");
    } finally {
      setIsLoading(false);
    }
  };

  const handleCreateRequest = (parentId: string | null = null) => {
    const newReq: RequestItem = {
      id: uuidv4(),
      name: 'New Request',
      method: 'GET',
      url: '',
      headers: [],
      bodyType: 'none',
      body: '',
      formData: [],
      parentId,
      type: 'request',
      createdAt: Date.now()
    };
    setItems(prev => [...prev, newReq]);
    setActiveId(newReq.id);
    setResponse(null);
  };

  const handleCreateFolder = (parentId: string | null = null) => {
    const newFolder: SidebarItem = {
      id: uuidv4(),
      name: 'New Folder',
      parentId,
      type: 'folder',
      createdAt: Date.now()
    };
    setItems(prev => [...prev, newFolder]);
  };

  const handleExportWorkspace = () => {
    const data = JSON.stringify({ version: 1, items }, null, 2);
    const blob = new Blob([data], { type: 'application/json' });
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `constrictor-workspace-${new Date().toISOString().split('T')[0]}.json`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const handleImportWorkspace = async (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;

    const reader = new FileReader();
    reader.onload = async (event) => {
      try {
        const content = event.target?.result as string;
        const parsed = JSON.parse(content);
        if (Array.isArray(parsed)) {
          if (confirm('Importing will overwrite your current workspace. Proceed?')) {
            setItems(parsed);
            setActiveId(null);
            setResponse(null);
          }
        } else if (parsed.items && Array.isArray(parsed.items)) {
          if (confirm('Importing will overwrite your current workspace. Proceed?')) {
            setItems(parsed.items);
            setActiveId(null);
            setResponse(null);
          }
        }
      } catch (err) {
        alert('Invalid workspace file.');
      }
    };
    reader.readAsText(file);
    e.target.value = '';
  };

  const handleDeleteItem = (id: string) => {
    setItems(prev => prev.filter(item => item.id !== id && item.parentId !== id));
    if (activeId === id) setActiveId(null);
  };

  const handleRenameItem = (id: string, newName: string) => {
    setItems(prev => prev.map(item => item.id === id ? { ...item, name: newName } : item));
  };

  const handleMoveItem = (itemId: string, targetId: string | null) => {
    if (itemId === targetId) return;
    const isDescendant = (descendantId: string, ancestorId: string, items: SidebarItem[]): boolean => {
      let current: SidebarItem | undefined = items.find(i => i.id === descendantId);
      while (current && current.parentId) {
        if (current.parentId === ancestorId) return true;
        current = items.find(i => i.id === current!.parentId);
      }
      return false;
    };
    setItems(prev => {
      if (targetId !== null && isDescendant(targetId, itemId, prev)) return prev;
      return prev.map(item => item.id === itemId ? { ...item, parentId: targetId } : item);
    });
  };

  return (
    <div className="flex h-screen w-full bg-[#131314] text-[#e8eaed] overflow-hidden font-sans">
      <Sidebar
        items={items}
        activeId={activeId}
        onSelect={setActiveId}
        onCreateRequest={handleCreateRequest}
        onCreateFolder={handleCreateFolder}
        onDelete={handleDeleteItem}
        onRename={handleRenameItem}
        onMoveItem={handleMoveItem}
        onOpenSettings={() => setIsSettingsOpen(true)}
        onExport={handleExportWorkspace}
        onImport={handleImportWorkspace}
        gdriveEnabled={settings.gdrive.enabled && !!settings.gdrive.accessToken}
        syncStatus={syncStatus}
      />

      <main className="flex flex-1 overflow-hidden">
        {activeItem ? (
          <>
            <div className="flex-[1.2] min-w-0 border-r border-[#3c4043] bg-[#131314]">
              <RequestEditor
                request={activeItem}
                onUpdate={updateActiveRequest}
                onSend={handleSendRequest}
                isLoading={isLoading}
              />
            </div>
            <div className="flex-1 min-w-0 bg-[#131314]">
              <ResponseViewer
                response={response}
                isLoading={isLoading}
                error={error}
              />
            </div>
          </>
        ) : (
          <div className="flex flex-col items-center justify-center w-full space-y-4 opacity-20 select-none">
            <svg className="w-32 h-32" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="0.5">
              <path strokeLinecap="round" strokeLinejoin="round" d="M13 10V3L4 14h7v7l9-11h-7z" />
            </svg>
            <p className="text-sm tracking-[0.5em] font-black uppercase heading-bold">Constrictor</p>
          </div>
        )}
      </main>

      {isSettingsOpen && (
        <SettingsModal
          settings={settings}
          onUpdate={setSettings}
          onClose={() => setIsSettingsOpen(false)}
        />
      )}
    </div>
  );
};

export default App;
