import React, { useState, useEffect } from 'react';
import Sidebar from './components/Sidebar';
import RequestEditor from './components/RequestEditor';
import ResponseViewer from './components/ResponseViewer';
import SettingsModal from './components/SettingsModal';
import { SidebarItem, RequestItem, ResponseData, AppSettings } from './types';
import { GetWorkspace, SaveWorkspace, ExecuteRequest } from './wails';
import { v4 as uuidv4 } from 'uuid';

const isRequestItem = (item: SidebarItem): item is RequestItem => {
  return item.type === 'request';
};

const App: React.FC = () => {
  const [items, setItems] = useState<SidebarItem[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);
  const [response, setResponse] = useState<ResponseData | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [isSettingsOpen, setIsSettingsOpen] = useState(false);
  const [settings, setSettings] = useState<AppSettings>({});
  const [isSettingsLoaded, setIsSettingsLoaded] = useState(false);

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
    setIsSettingsLoaded(true);
  }, []);

  // Save settings to localStorage
  useEffect(() => {
    if (!isSettingsLoaded) return;
    localStorage.setItem('constrictor_settings', JSON.stringify(settings));
  }, [settings, isSettingsLoaded]);

  // Load workspace from backend
  useEffect(() => {
    const loadWorkspace = async () => {
      try {
        const items = await GetWorkspace();

        if (items && items.length > 0) {
          // Normalize items to ensure all required fields are present
          const normalizedItems = items.map((item: any) => {
            if (item.type === 'request') {
              return {
                ...item,
                headers: item.headers || [],
                formData: item.formData || [],
                body: item.body || '',
                bodyType: item.bodyType || 'none',
                method: item.method || 'GET',
                url: item.url || '',
              };
            }
            return item;
          });
          setItems(normalizedItems);
          if (normalizedItems[0].type === 'request') {
            setActiveId(normalizedItems[0].id);
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

  // Save workspace to backend when items change
  useEffect(() => {
    if (items.length === 0) return;

    const saveWorkspace = async () => {
      try {
        await SaveWorkspace(items);
      } catch (err) {
        console.error('Failed to save workspace:', err);
      }
    };

    const timeoutId = setTimeout(saveWorkspace, 500); // Debounce
    return () => clearTimeout(timeoutId);
  }, [items]);

  const activeItem = items.find(i => i.id === activeId);

  const updateActiveRequest = (updates: Partial<RequestItem>) => {
    if (!activeId) return;
    const item = items.find(i => i.id === activeId);
    if (!item || !isRequestItem(item)) return;
    setItems(prev => prev.map(item => item.id === activeId ? { ...item, ...updates } : item));
  };

  const handleSendRequest = async () => {
    if (!activeItem) return;
    setIsLoading(true);
    setError(null);
    setResponse(null);

    try {
      const response = await ExecuteRequest({
        method: activeItem.method,
        url: activeItem.url,
        headers: activeItem.headers,
        bodyType: activeItem.bodyType,
        body: activeItem.body,
        formData: activeItem.formData
      });

      setResponse(response);
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
        let itemsToImport: any[] = [];

        if (Array.isArray(parsed)) {
          itemsToImport = parsed;
        } else if (parsed.items && Array.isArray(parsed.items)) {
          itemsToImport = parsed.items;
        } else {
          alert('Invalid workspace file: missing items array.');
          return;
        }

        const isValidItem = (item: any): boolean => {
          return (
            item &&
            typeof item.id === 'string' &&
            typeof item.type === 'string' &&
            typeof item.name === 'string' &&
            typeof item.createdAt === 'number'
          );
        };

        if (!itemsToImport.every(isValidItem)) {
          console.error('Validation failed for imported items:', itemsToImport);
          alert('Invalid workspace file: one or more items missing required properties (id, type, name, createdAt).');
          return;
        }

          if (confirm('Importing will overwrite your current workspace. Proceed?')) {
          setItems(itemsToImport);
            setActiveId(null);
            setResponse(null);
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
      />

      <main className="flex flex-1 overflow-hidden">
        {activeItem && isRequestItem(activeItem) ? (
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
