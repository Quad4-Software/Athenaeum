import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/svelte";
import userEvent from "@testing-library/user-event";
import ErrorView from "./ErrorView.svelte";

vi.mock("$lib/router.svelte", () => ({
  router: { navigate: vi.fn() },
}));

vi.mock("$lib/stores/auth.svelte", () => ({
  auth: { authEnabled: false },
}));

vi.mock("$lib/stores/i18n.svelte", () => ({
  i18n: {
    t: (key: string) =>
      (
        ({
          "error.notFoundTitle": "Page not found",
          "error.notFoundMessage": "The page or resource you requested does not exist.",
          "error.crashTitle": "Something went wrong",
          "error.crashMessage": "The app hit an unexpected error. You can try again.",
          "error.retry": "Retry",
          "app.goToLibrary": "Go to library",
        }) as Record<string, string>
      )[key] ?? key,
  },
}));

describe("ErrorView", () => {
  it("calls onRetry when Retry is clicked", async () => {
    const user = userEvent.setup();
    const onRetry = vi.fn();
    render(ErrorView, {
      props: {
        title: "Something went wrong",
        message: "The app hit an unexpected error. You can try again.",
        onRetry,
      },
    });
    await user.click(screen.getByRole("button", { name: "Retry" }));
    expect(onRetry).toHaveBeenCalledTimes(1);
  });
});
