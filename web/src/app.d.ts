/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />

interface ImportMetaEnv {
  readonly VITE_DEMO?: string;
  readonly VITE_SLIM?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}

declare module "*.css";

declare module "*?url" {
  const url: string;
  export default url;
}
