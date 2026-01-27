/**
 * Wails TypeScript bindings for Constrictor REST Client
 * 
 * This module provides TypeScript definitions and wrapper functions for Wails runtime bindings.
 * It allows the frontend to communicate with the Go backend without HTTP requests.
 */

import { SidebarItem, ResponseData } from './types';

/**
 * ExecuteRequestInput matches the Go ExecuteRequestInput struct
 */
export interface ExecuteRequestInput {
  method: string;
  url: string;
  headers: Array<{
    key: string;
    value: string;
    enabled: boolean;
  }>;
  bodyType: 'none' | 'json' | 'form-data' | 'url-encoded';
  body: string;
  formData: Array<{
    key: string;
    value: string;
    enabled: boolean;
  }>;
}

/**
 * ExecutionResult matches the Go executor.ExecutionResult struct
 */
export interface ExecutionResult {
  status: number;
  statusText: string;
  headers: Record<string, string>;
  body: string;
  timeMs: number;
  sizeBytes: number;
  error?: {
    message: string;
    type: string;
  };
}

/**
 * App interface matching the Go App struct methods
 */
interface App {
  GetWorkspace(): Promise<SidebarItem[]>;
  SaveWorkspace(items: SidebarItem[]): Promise<void>;
  ExecuteRequest(req: ExecuteRequestInput): Promise<ExecutionResult>;
}

/**
 * Check if Wails runtime is available
 */
export function isWailsRuntime(): boolean {
  return typeof window !== 'undefined' && 
         typeof (window as any).go !== 'undefined' && 
         typeof (window as any).go.main !== 'undefined' &&
         typeof (window as any).go.main.App !== 'undefined';
}

/**
 * Get the Wails App instance
 */
function getApp(): App | null {
  if (!isWailsRuntime()) {
    return null;
  }
  return (window as any).go.main.App as App;
}

/**
 * Get workspace from the backend
 * @returns Promise resolving to array of SidebarItem
 * @throws Error if Wails runtime is not available or request fails
 */
export async function GetWorkspace(): Promise<SidebarItem[]> {
  const app = getApp();
  if (!app) {
    throw new Error('Wails runtime is not available. This app must be run in Wails.');
  }
  try {
    return await app.GetWorkspace();
  } catch (error: any) {
    throw new Error(`Failed to load workspace: ${error.message || error}`);
  }
}

/**
 * Save workspace to the backend
 * @param items Array of SidebarItem to save
 * @throws Error if Wails runtime is not available or save fails
 */
export async function SaveWorkspace(items: SidebarItem[]): Promise<void> {
  const app = getApp();
  if (!app) {
    throw new Error('Wails runtime is not available. This app must be run in Wails.');
  }
  try {
    await app.SaveWorkspace(items);
  } catch (error: any) {
    throw new Error(`Failed to save workspace: ${error.message || error}`);
  }
}

/**
 * Execute an HTTP request
 * @param req ExecuteRequestInput containing request details
 * @returns Promise resolving to ExecutionResult
 * @throws Error if Wails runtime is not available or execution fails
 */
export async function ExecuteRequest(req: ExecuteRequestInput): Promise<ResponseData> {
  const app = getApp();
  if (!app) {
    throw new Error('Wails runtime is not available. This app must be run in Wails.');
  }
  try {
    const result = await app.ExecuteRequest(req);
    
    // Convert ExecutionResult to ResponseData format expected by frontend
    if (result.error) {
      throw new Error(result.error.message || 'Request failed');
    }
    
    return {
      status: result.status,
      statusText: result.statusText,
      headers: result.headers,
      body: result.body,
      time: result.timeMs,
      size: result.sizeBytes,
    };
  } catch (error: any) {
    // If it's already an Error with a message, rethrow it
    if (error instanceof Error) {
      throw error;
    }
    // Otherwise wrap it
    throw new Error(`Failed to execute request: ${error.message || error}`);
  }
}
