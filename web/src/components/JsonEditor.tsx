import React, { useState, useEffect } from 'react';
import Editor from 'react-simple-code-editor';
import { highlight, languages } from 'prismjs';
import 'prismjs/components/prism-json';
import 'prismjs/themes/prism-tomorrow.css';

interface JsonEditorProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  error?: string | null;
}

const JsonEditor: React.FC<JsonEditorProps> = ({ value, onChange, placeholder, error }) => {
  const [isInvalid, setIsInvalid] = useState(false);

  useEffect(() => {
    if (value.trim()) {
      try {
        JSON.parse(value);
        setIsInvalid(false);
      } catch {
        setIsInvalid(true);
      }
    } else {
      setIsInvalid(false);
    }
  }, [value]);

  return (
    <div className={`flex-1 flex flex-col bg-[#1e1e20] rounded-xl border transition-colors ${
      isInvalid || error ? 'border-[#f28b82]' : 'border-[#3c4043]'
    } focus-within:border-[#8ab4f8] overflow-hidden`}>
      {isInvalid && (
        <div className="bg-[#2a1a1a] px-4 py-2 text-[#f28b82] text-[11px] font-bold uppercase tracking-widest border-b border-[#f28b82]/20">
          Invalid JSON
        </div>
      )}
      <Editor
        value={value}
        onValueChange={onChange}
        highlight={code => highlight(code || '', languages.json, 'json')}
        placeholder={placeholder}
        className="flex-1 font-mono text-[14px] leading-relaxed"
        style={{
          fontFamily: '"JetBrains Mono", "Fira Code", monospace',
          fontSize: '14px',
          lineHeight: '1.6',
          minHeight: '100%'
        }}
        padding={{
          top: 16,
          bottom: 16,
          left: 20,
          right: 20
        }}
      />
    </div>
  );
};

export default JsonEditor;
