import React, { useState, useMemo, useRef } from 'react';
import { SidebarItem, RequestItem } from '../types';
import MethodBadge from './MethodBadge';

interface SidebarProps {
  items: SidebarItem[];
  activeId: string | null;
  onSelect: (id: string) => void;
  onCreateRequest: (parentId?: string | null) => void;
  onCreateFolder: (parentId?: string | null) => void;
  onDelete: (id: string) => void;
  onRename: (id: string, newName: string) => void;
  onMoveItem: (itemId: string, targetId: string | null) => void;
  onOpenSettings: () => void;
  onExport: () => void;
  onImport: (e: React.ChangeEvent<HTMLInputElement>) => void;
}

const Sidebar: React.FC<SidebarProps> = ({ 
  items, 
  activeId, 
  onSelect, 
  onCreateRequest, 
  onCreateFolder,
  onDelete,
  onRename,
  onMoveItem,
  onOpenSettings,
  onExport,
  onImport
}) => {
  const [editingId, setEditingId] = useState<string | null>(null);
  const [tempName, setTempName] = useState('');
  const [dragOverId, setDragOverId] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');
  const fileInputRef = useRef<HTMLInputElement>(null);

  const handleDragStart = (e: React.DragEvent, id: string) => {
    e.dataTransfer.setData('text/plain', id);
    e.dataTransfer.effectAllowed = 'move';
  };

  const handleDragOver = (e: React.DragEvent, id: string | null) => {
    e.preventDefault();
    if (id !== dragOverId) setDragOverId(id);
  };

  const handleDrop = (e: React.DragEvent, targetId: string | null) => {
    e.preventDefault();
    setDragOverId(null);
    const itemId = e.dataTransfer.getData('text/plain');
    if (itemId) onMoveItem(itemId, targetId);
  };

  const filteredItems = useMemo(() => {
    const query = searchQuery.toLowerCase().trim();
    if (!query) return items;
    return items.filter(item => item.name.toLowerCase().includes(query));
  }, [items, searchQuery]);

  const renderItem = (item: SidebarItem) => {
    const isActive = activeId === item.id;
    const isFolder = item.type === 'folder';
    const isBeingDraggedOver = dragOverId === item.id && isFolder;

    const handleEditStart = (e: React.MouseEvent) => {
      e.stopPropagation();
      setEditingId(item.id);
      setTempName(item.name);
    };

    const handleEditEnd = () => {
      onRename(item.id, tempName);
      setEditingId(null);
    };

    return (
      <div 
        key={item.id}
        draggable={item.type === 'request'}
        onDragStart={(e) => handleDragStart(e, item.id)}
        onDragOver={(e) => isFolder ? handleDragOver(e, item.id) : null}
        onDragLeave={() => isFolder ? setDragOverId(null) : null}
        onDrop={(e) => isFolder ? handleDrop(e, item.id) : null}
        onClick={() => !isFolder && onSelect(item.id)}
        className={`group relative flex items-center px-4 py-2 text-[13px] cursor-pointer transition-all duration-150 border-l-4 ${
          isActive 
            ? 'bg-[#2a2b2f] border-[#8ab4f8] text-[#8ab4f8] font-bold' 
            : 'border-transparent hover:bg-[#2a2b2f] text-[#9aa0a6] hover:text-[#e8eaed]'
        } ${isBeingDraggedOver ? 'bg-[#3c4043] scale-[1.02]' : ''}`}
      >
        <div className="flex items-center gap-3 w-full min-w-0">
          {isFolder ? (
            <svg className="w-4 h-4 text-[#9aa0a6] shrink-0" fill="currentColor" viewBox="0 0 20 20">
              <path d="M2 6a2 2 0 012-2h5l2 2h5a2 2 0 012 2v6a2 2 0 01-2 2H4a2 2 0 01-2-2V6z" />
            </svg>
          ) : (
            <MethodBadge method={(item as RequestItem).method} />
          )}

          {editingId === item.id ? (
            <input 
              autoFocus
              className="bg-[#131314] text-white rounded px-1 w-full outline-none ring-1 ring-[#8ab4f8] font-bold"
              value={tempName}
              onChange={(e) => setTempName(e.target.value)}
              onBlur={handleEditEnd}
              onKeyDown={(e) => e.key === 'Enter' && handleEditEnd()}
            />
          ) : (
            <span className="truncate flex-1 tracking-tight font-medium">{item.name}</span>
          )}
        </div>

        <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1 transition-opacity">
          <button onClick={handleEditStart} className="p-1 hover:text-[#e8eaed] text-[#5f6368]" title="Rename">
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M15.232 5.232l3.536 3.536m-2.036-5.036a2.5 2.5 0 113.536 3.536L6.5 21.036H3v-3.572L16.732 3.732z" strokeWidth="2"/></svg>
          </button>
          <button onClick={(e) => { e.stopPropagation(); onDelete(item.id); }} className="p-1 hover:text-[#f28b82] text-[#5f6368]" title="Delete">
            <svg className="w-3.5 h-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" strokeWidth="2"/></svg>
          </button>
        </div>
      </div>
    );
  };

  const isSearching = searchQuery.trim().length > 0;
  const rootItems = isSearching ? filteredItems : items.filter(i => !i.parentId);

  return (
    <div className="w-72 flex flex-col border-r border-[#3c4043] bg-[#1e1e20] select-none" onDragOver={(e) => handleDragOver(e, null)} onDrop={(e) => handleDrop(e, null)}>
      <div className="p-5 flex items-center justify-between">
        <h1 className="heading-bold text-sm tracking-widest text-[#e8eaed]">CONSTRICTOR</h1>
        <div className="flex gap-1">
          <button onClick={() => onCreateFolder()} className="p-2 rounded hover:bg-[#3c4043] text-[#9aa0a6] hover:text-[#e8eaed]" title="New Folder">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeWidth="2" d="M9 13h6m-3-3v6m-9 1V7a2 2 0 012-2h6l2 2h6a2 2 0 012 2v8a2 2 0 01-2 2H5a2 2 0 01-2-2z" /></svg>
          </button>
          <button onClick={() => onCreateRequest()} className="p-2 rounded hover:bg-[#3c4043] text-[#9aa0a6] hover:text-[#e8eaed]" title="New Request">
            <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M12 4v16m8-8H4" /></svg>
          </button>
        </div>
      </div>

      <div className="px-4 pb-4">
        <div className="relative group">
          <input 
            type="text" 
            placeholder="Search..."
            className="w-full bg-[#131314] border border-[#3c4043] rounded-lg py-2 pl-9 pr-3 text-[13px] text-[#e8eaed] placeholder-[#5f6368] outline-none focus:border-[#8ab4f8] transition-colors"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
          />
          <svg className="absolute left-3 top-2.5 w-4 h-4 text-[#5f6368] group-focus-within:text-[#8ab4f8] transition-colors" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2" d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
      </div>

      <div className="flex-1 overflow-y-auto py-2 space-y-1">
        {rootItems.map(item => {
          if (!isSearching && item.type === 'folder') {
            const children = items.filter(i => i.parentId === item.id);
            return (
              <div key={item.id}>
                {renderItem(item)}
                <div className="ml-6 border-l border-[#3c4043] my-1 space-y-1">
                  {children.map(child => renderItem(child))}
                  {children.length === 0 && <div className="py-1 px-3 text-[11px] text-[#5f6368] italic font-bold">EMPTY</div>}
                </div>
              </div>
            );
          }
          return renderItem(item);
        })}
      </div>

      <div className="p-4 border-t border-[#3c4043] flex flex-col gap-3">
        <div className="flex gap-2">
           <button 
            onClick={onExport}
            className="flex-1 flex items-center justify-center gap-2 py-2 text-[11px] font-bold uppercase tracking-widest text-[#9aa0a6] hover:text-[#e8eaed] bg-[#2a2b2f] rounded transition-colors"
          >
            Export
          </button>
          <button 
            onClick={() => fileInputRef.current?.click()}
            className="flex-1 flex items-center justify-center gap-2 py-2 text-[11px] font-bold uppercase tracking-widest text-[#9aa0a6] hover:text-[#e8eaed] bg-[#2a2b2f] rounded transition-colors"
          >
            Import
          </button>
          <input type="file" ref={fileInputRef} onChange={onImport} accept=".json" className="hidden" />
        </div>
        <button 
          onClick={onOpenSettings}
          className="flex items-center gap-3 text-[12px] font-bold uppercase tracking-widest text-[#9aa0a6] hover:text-[#8ab4f8] transition-colors"
        >
          <svg className="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" strokeWidth="2"/><path d="M15 12a3 3 0 11-6 0 3 3 0 016 0z" strokeWidth="2"/></svg>
          Settings
        </button>
        <div className="text-[10px] text-[#5f6368] uppercase tracking-widest font-bold flex justify-between">
          <span>v1.3.1</span>
          <span>Workspace Sync</span>
        </div>
      </div>
    </div>
  );
};

export default Sidebar;
