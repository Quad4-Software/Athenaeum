/**
 * Shared viewport breakpoints.
 *
 * JS side: use these media query strings with `new MediaQuery(...)`.
 * CSS side: use `theme(--breakpoint-*)` in @media rules (Tailwind v4 resolves
 * it); `--breakpoint-xs` is declared in app.css, sm/md/lg are Tailwind
 * defaults. Keep the values below in sync with those.
 */
export const BP_XS_PX = 480;
export const BP_SM_PX = 640;
export const BP_MD_PX = 768;
export const BP_LG_PX = 1024;

/** Viewports at or above the xs breakpoint. */
export const MQ_XS_UP = `(min-width: ${BP_XS_PX}px)`;
/** Viewports below the md breakpoint (phones). */
export const MQ_BELOW_MD = `(max-width: ${BP_MD_PX - 1}px)`;
/** Viewports at or above the md breakpoint (tablets and up). */
export const MQ_MD_UP = `(min-width: ${BP_MD_PX}px)`;
