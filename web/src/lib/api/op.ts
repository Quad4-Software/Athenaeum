import { apiOp, apiPath, type ApiOperationId } from "./generated/paths";

/** Resolve a generated operation id to a concrete URL (fills `{param}` segments). */
export function opURL(id: ApiOperationId, params: Record<string, string | number> = {}): string {
  return apiPath(apiOp(id).path, params);
}
