import React, { useState } from 'react';
import { ResponseData } from '../types';

interface ResponseViewerProps {
  response: ResponseData | null;
  isLoading: boolean;
  error: string | null;
}

const JsonNode: React.FC<{ data: any; label?: string; depth?: number }> = ({ data, label, depth = 0 }) => {
  const [isCollapsed, setIsCollapsed] = useState(false);
  const isObject = data !== null && typeof data === 'object';
  const isArray = Array.isArray(data);

  const renderValue = (val: any) => {
    if (typeof val === 'string') return <span className="text-[#81c995]">"{val}"</span>;
    if (typeof val === 'number') return <span className="text-[#8ab4f8]">{val}</span>;
    if (typeof val === 'boolean') return <span className="text-[#fdd663]">{val.toString()}</span>;
    if (val === null) return <span className="text-[#9aa0a6] italic">null</span>;
    return null;
  };

  if (!isObject) {
    return (
      <div className="flex gap-2 py-0.5 text-[14px]">
        {label && <span className="text-[#9aa0a6] font-bold">{label}:</span>}
        {renderValue(data)}
      </div>
    );
  }

  const entries = isArray ? data : Object.entries(data);
  const count = entries.length;

  return (
    <div className="flex flex-col">
      <div
        onClick={() => setIsCollapsed(!isCollapsed)}
        className="flex items-center gap-2 py-1 cursor-pointer hover:bg-[#2a2b2f] group rounded px-1 transition-colors"
      >
        <span className={`text-[10px] text-[#5f6368] transition-transform ${isCollapsed ? '' : 'rotate-90'}`}>▶</span>
        {label && <span className="text-[#e8eaed] font-bold text-[14px]">{label}:</span>}
        <span className="text-[#5f6368] text-[12px] font-bold">
          {isArray ? `Array(${count})` : `Object(${count})`}
        </span>
      </div>
      {!isCollapsed && (
        <div className="ml-4 pl-3 border-l border-[#3c4043]">
          {entries.map((entry: any, i: number) => {
            const nodeLabel = isArray ? i.toString() : entry[0];
            const nodeData = isArray ? entry : entry[1];
            return <JsonNode key={i} label={nodeLabel} data={nodeData} depth={depth + 1} />;
          })}
        </div>
      )}
    </div>
  );
};

const ResponseViewer: React.FC<ResponseViewerProps> = ({ response, isLoading, error }) => {
  const [activeTab, setActiveTab] = useState<'body' | 'headers'>('body');

  if (isLoading) {
    return (
      <div className="h-full flex flex-col items-center justify-center space-y-6 bg-[#131314]">
        <div className="w-14 h-14 border-4 border-[#3c4043] border-t-[#8ab4f8] rounded-full animate-spin"></div>
        <p className="heading-bold text-[12px] text-[#9aa0a6] tracking-[0.4em]">Processing...</p>
      </div>
    );
  }

  if (error) {
    return (
      <div className="h-full p-8 flex flex-col items-center justify-center text-center bg-[#131314]">
        <div className="bg-[#2a1a1a] text-[#f28b82] p-10 rounded-2xl border border-[#f28b82]/20 max-w-sm shadow-xl">
          <h3 className="heading-bold text-sm mb-4">Request Failure</h3>
          <p className="text-[14px] font-mono opacity-80 leading-relaxed">{error}</p>
        </div>
      </div>
    );
  }

  if (!response) {
    return (
      <div className="h-full flex flex-col items-center justify-center opacity-10 bg-[#131314]">
        <p className="heading-bold text-[14px] tracking-[0.6em]">NOVA IDLE</p>
      </div>
    );
  }

  let parsedBody: any = null;
  try {
    parsedBody = JSON.parse(response.body);
  } catch {
    parsedBody = response.body;
  }

  return (
    <div className="h-full flex flex-col bg-[#131314] overflow-hidden">
      <div className="flex-shrink-0 p-6 border-b border-[#3c4043] flex items-center justify-between">
        <div className="flex items-center gap-10">
          <div className="flex flex-col">
            <span className="text-[10px] text-[#5f6368] font-bold uppercase tracking-widest mb-1">Status</span>
            <span className={`text-[15px] font-bold ${response.status < 400 ? 'text-[#81c995]' : 'text-[#f28b82]'}`}>{response.status}</span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] text-[#5f6368] font-bold uppercase tracking-widest mb-1">Time</span>
            <span className="text-[15px] font-bold text-[#8ab4f8]">{response.time}ms</span>
          </div>
          <div className="flex flex-col">
            <span className="text-[10px] text-[#5f6368] font-bold uppercase tracking-widest mb-1">Size</span>
            <span className="text-[15px] font-bold text-[#e8eaed]">{(response.size / 1024).toFixed(2)}kb</span>
          </div>
        </div>
      </div>

      <div className="flex-shrink-0 flex border-b border-[#3c4043] px-6 gap-8">
        <button onClick={() => setActiveTab('body')} className={`py-4 text-[12px] font-bold uppercase tracking-widest border-b-2 transition-all ${activeTab === 'body' ? 'border-[#8ab4f8] text-[#8ab4f8]' : 'border-transparent text-[#9aa0a6] hover:text-[#e8eaed]'}`}>Response Body</button>
        <button onClick={() => setActiveTab('headers')} className={`py-4 text-[12px] font-bold uppercase tracking-widest border-b-2 transition-all ${activeTab === 'headers' ? 'border-[#8ab4f8] text-[#8ab4f8]' : 'border-transparent text-[#9aa0a6] hover:text-[#e8eaed]'}`}>Headers</button>
      </div>

      <div className="flex-1 min-h-0 overflow-y-auto p-6">
        {activeTab === 'body' ? (
          <div className="bg-[#1e1e20] p-6 rounded-xl border border-[#3c4043] font-mono">
            {typeof parsedBody === 'object' ? (
              <JsonNode data={parsedBody} />
            ) : (
              <pre className="whitespace-pre-wrap text-[14px] text-[#e8eaed] leading-relaxed break-words overflow-wrap-anywhere">{parsedBody}</pre>
            )}
          </div>
        ) : (
          <div className="space-y-1 font-mono text-[13px]">
            {Object.entries(response.headers).map(([key, value]) => (
              <div key={key} className="flex border-b border-[#3c4043]/50 py-3 group hover:bg-[#1e1e20] px-2 transition-colors">
                <span className="w-1/3 font-bold text-[#9aa0a6] select-all uppercase tracking-tighter text-[11px] self-center flex-shrink-0">{key}</span>
                <span className="flex-1 text-[#e8eaed] select-all break-words overflow-wrap-anywhere min-w-0">{value}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ResponseViewer;
