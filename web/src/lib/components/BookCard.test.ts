import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/svelte";
import BookCard from "./BookCard.svelte";
import type { Book } from "$lib/api/types";

const book: Book = {
  id: 7,
  title: "The Left Hand of Darkness",
  author: "Ursula K. Le Guin",
  format: "epub",
  relPath: "leguin/left-hand.epub",
  fileSize: 1024,
  hasCover: false,
  addedAt: new Date().toISOString(),
  modifiedAt: new Date().toISOString(),
};

describe("BookCard", () => {
  it("renders the title and author", () => {
    render(BookCard, { props: { book } });
    expect(screen.getByTitle(book.title)).toHaveTextContent(book.title);
    expect(screen.getByText(book.author)).toBeInTheDocument();
  });

  it("links to the book detail page", () => {
    render(BookCard, { props: { book } });
    const link = screen.getByRole("link");
    expect(link).toHaveAttribute("href", "/book/7");
  });
});
