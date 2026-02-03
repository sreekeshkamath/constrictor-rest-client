export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'OPTIONS' | 'HEAD';
export type BodyType = 'none' | 'json' | 'form-data' | 'url-encoded';
export type AuthType = 'none' | 'inherit' | 'bearer' | 'basic' | 'apikey' | 'oauth2' | 'oauth1' | 'digest' | 'ntlm' | 'aws' | 'hawk' | 'asap' | 'netrc';

export interface Header {
  key: string;
  value: string;
  enabled: boolean;
}

export interface FormDataItem {
  key: string;
  value: string;
  enabled: boolean;
}

export interface AuthConfig {
  type: AuthType;
  config: Record<string, any>;
}

export interface RequestItem {
  id: string;
  name: string;
  method: HttpMethod;
  url: string;
  headers: Header[];
  bodyType: BodyType;
  body: string;
  formData: FormDataItem[];
  auth?: AuthConfig;
  parentId?: string | null;
  type: 'request';
  createdAt: number;
}

export interface FolderItem {
  id: string;
  name: string;
  parentId?: string | null;
  type: 'folder';
  createdAt: number;
  auth?: AuthConfig;
}

export type SidebarItem = RequestItem | FolderItem;

export interface ResponseData {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body: string;
  time: number;
  size: number;
}

export interface AppSettings {
  // Reserved for future settings
}
