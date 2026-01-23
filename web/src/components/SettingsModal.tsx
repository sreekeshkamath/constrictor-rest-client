import React from 'react';
import { AppSettings } from '../types';

interface SettingsModalProps {
  settings: AppSettings;
  onUpdate: (settings: AppSettings) => void;
  onClose: () => void;
}

const SettingsModal: React.FC<SettingsModalProps> = ({ settings, onUpdate, onClose }) => {


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
