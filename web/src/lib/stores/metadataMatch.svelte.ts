import { api } from "$lib/api/client";
import { createJobPoller } from "$lib/stores/job-poller.svelte";

export const metadataMatch = createJobPoller({
  fetchStatus: () => api.metadataMatchStatus(),
  isActive: (st) => st.running,
});
