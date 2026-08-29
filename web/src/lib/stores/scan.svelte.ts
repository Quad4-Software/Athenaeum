import { api } from "$lib/api/client";
import { createJobPoller } from "$lib/stores/job-poller.svelte";

export const scan = createJobPoller({
  fetchStatus: () => api.scanStatus(),
  isActive: (st) => st.scanning,
});
