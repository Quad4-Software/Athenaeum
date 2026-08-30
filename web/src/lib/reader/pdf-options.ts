import type { DocumentInitParameters } from "pdfjs-dist/types/src/display/api";

const pdfAssetBase = "/pdfjs/";

/** Open options for the canvas PDF reader (no scripting / XFA). */
export function pdfOpenOptions(url: string): DocumentInitParameters {
  return {
    url,
    withCredentials: true,
    wasmUrl: `${pdfAssetBase}wasm/`,
    standardFontDataUrl: `${pdfAssetBase}standard_fonts/`,
    cMapUrl: `${pdfAssetBase}cmaps/`,
    iccUrl: `${pdfAssetBase}iccs/`,
    enableXfa: false,
  };
}
