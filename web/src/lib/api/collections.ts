import type { Collection, CollectionKind, SmartQuery } from "./types";
import { request } from "./core";
import { opURL } from "./op";

export const collectionsApi = {
  listCollections: () => request<Collection[]>(opURL("GET__api_collections")),

  createCollection: (
    name: string,
    description = "",
    kind: CollectionKind = "manual",
    query?: SmartQuery,
  ) =>
    request<Collection>(opURL("POST__api_collections"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ name, description, kind, query }),
    }),

  deleteCollection: (id: number) =>
    request<void>(opURL("DELETE__api_collections__id", { id }), { method: "DELETE" }),

  addToCollection: (collectionId: number, bookId: number) =>
    request<void>(opURL("POST__api_collections__id__books__bookId", { id: collectionId, bookId }), {
      method: "POST",
    }),

  removeFromCollection: (collectionId: number, bookId: number) =>
    request<void>(
      opURL("DELETE__api_collections__id__books__bookId", { id: collectionId, bookId }),
      { method: "DELETE" },
    ),

  listFavorites: () => request<{ ids: number[] }>(opURL("GET__api_favorites")),

  setFavorite: (bookId: number, favorite: boolean) =>
    request<{ favorite: boolean }>(opURL("PUT__api_books__id__favorite", { id: bookId }), {
      method: "PUT",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ favorite }),
    }),
};
