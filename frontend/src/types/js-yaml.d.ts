declare module 'js-yaml' {
  export function load(input: string, options?: any): any;
  export function dump(obj: any, options?: any): string;
  export function loadAll(input: string, callback?: (doc: any) => void, options?: any): any[];
  export function safeLoad(input: string, options?: any): any;
  export function safeDump(obj: any, options?: any): string;
} 