import type { SelectedTypes } from "@/types/media-and-posters/media-item-and-library";

export type TYPE_MEDIA_SCOPE_OPTIONS = "all" | "movies" | "shows" | "collections";
export const MEDIA_SCOPE_OPTIONS: { value: TYPE_MEDIA_SCOPE_OPTIONS; label: string }[] = [
  { value: "all", label: "All" },
  { value: "movies", label: "Movies" },
  { value: "shows", label: "Shows" },
  { value: "collections", label: "Collections" },
];

export type TYPE_PRIORITY_OPTIONS = 1 | 2 | 3 | 4 | 5;
export const PRIORITY_OPTIONS: { value: TYPE_PRIORITY_OPTIONS; label: string }[] = [
  { value: 1, label: "1 (Highest)" },
  { value: 2, label: "2" },
  { value: 3, label: "3" },
  { value: 4, label: "4" },
  { value: 5, label: "5 (Lowest)" },
];

export interface UserSubscription {
  id: number;
  username: string;
  creator_id: string;
  image_types: SelectedTypes;
  priority: TYPE_PRIORITY_OPTIONS;
  media_scope: TYPE_MEDIA_SCOPE_OPTIONS;
  library_section: string | null;
  enabled: boolean;
  date_created: string;
  date_updated: string;
}

export interface CreateSubscriptionRequest {
  username: string;
  creator_id?: string;
  image_types: SelectedTypes;
  priority: TYPE_PRIORITY_OPTIONS;
  media_scope: TYPE_MEDIA_SCOPE_OPTIONS;
  library_section?: string | null;
  enabled: boolean;
}

export interface UpdateSubscriptionRequest {
  creator_id?: string;
  image_types: SelectedTypes;
  priority: TYPE_PRIORITY_OPTIONS;
  media_scope: TYPE_MEDIA_SCOPE_OPTIONS;
  library_section?: string | null;
  enabled: boolean;
}
