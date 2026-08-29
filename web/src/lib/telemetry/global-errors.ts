import { captureException } from "./sentry";

let registered = false;

/** Register window error handlers once. Safe to call before Sentry init. */
export function registerGlobalErrorHandlers() {
  if (registered || typeof window === "undefined") return;
  registered = true;

  window.addEventListener("error", (event) => {
    const err = event.error ?? event.message;
    console.error(err);
    captureException(err);
  });

  window.addEventListener("unhandledrejection", (event) => {
    console.error(event.reason);
    captureException(event.reason);
  });
}
