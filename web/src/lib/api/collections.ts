import type { Collection, CollectionKind, SmartQuery } from "./types";
import { request } from "./core";

export const collectionsApi = {
  listCollections: () => request<Collection[]>("/api/collections"),

  createCollection: (
    name: string,
    description = "",
    kind: CollectionKind = "manual",
    query?: SmartQuery,
  ) =>
    request<Collection>("/api/collections", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, description, kind, query }),
    }),

  deleteCollection: (id: number) => request<void>(`/api/collections/${id}`, { method: "DELETE" }),

  addToCollection: (collectionId: number, bookId: number) =>
    request<void>(`/api/collections/${collectionId}/books/${bookId}`, { method: "POST" }),

  removeFromCollection: (collectionId: number, bookId: number) =>
    request<void>(`/api/collections/${collectionId}/books/${bookId}`, { method: "DELETE" }),

  listFavorites: () => request<{ ids: number[] }>("/api/favorites"),

  setFavorite: (bookId: number, favorite: boolean) =>
    request<{ favorite: boolean }>(`/api/books/${bookId}/favorite`, {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ favorite }),
    }),
};
