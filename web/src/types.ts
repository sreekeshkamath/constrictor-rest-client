export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'OPTIONS' | 'HEAD';
export type BodyType = 'none' | 'json' | 'form-data' | 'url-encoded';

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

export interface RequestItem {
  id: string;
  name: string;
  method: HttpMethod;
  url: string;
  headers: Header[];
  bodyType: BodyType;
  body: string;
  formData: FormDataItem[];
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

export interface GDriveSettings {
  enabled: boolean;
  apiKey: string;
  clientId: string;
  accessToken?: string;
  lastSync?: string;
}

export interface AppSettings {
  gdrive: GDriveSettings;
}
