import React from 'react';
import { HttpMethod } from '../types';

interface MethodBadgeProps {
  method: HttpMethod;
  className?: string;
}

const MethodBadge: React.FC<MethodBadgeProps> = ({ method, className = "" }) => {
  const styles: Record<HttpMethod, string> = {
    GET: 'text-[#81c995]',
    POST: 'text-[#fdd663]',
    PUT: 'text-[#8ab4f8]',
    PATCH: 'text-[#c58af9]',
    DELETE: 'text-[#f28b82]',
    OPTIONS: 'text-[#9aa0a6]',
    HEAD: 'text-[#9aa0a6]'
  };

  return (
    <span className={`text-[11px] font-bold w-12 text-left shrink-0 uppercase tracking-tight ${styles[method]} ${className}`}>
      {method}
    </span>
  );
};

export default MethodBadge;
