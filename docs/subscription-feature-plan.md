# Implementation Plan: User Subscription Feature

## Overview

Allow users to "subscribe" to a MediUX creator, so all their future sets (matching selected image types and media categories) are automatically downloaded to the user's media server and tracked like existing saved sets.

---

## 1. Data Model

### New Table: `UserSubscriptions`

```sql
CREATE TABLE IF NOT EXISTS UserSubscriptions (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    username        TEXT NOT NULL,          -- subscribed creator's MediUX username
    creator_id      TEXT NOT NULL,          -- subscribed creator's MediUX user ID
    image_types     TEXT NOT NULL,          -- JSON: {"poster":true,"backdrop":false,"season_poster":true,"special_season_poster":false,"titlecard":true}
    media_scope     TEXT NOT NULL DEFAULT 'all',  -- 'all', 'movies', 'shows', 'collections'
    library_section TEXT,                   -- which of the user's library sections to match against (NULL = all)
    enabled         INTEGER NOT NULL DEFAULT 1,
    date_created    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    date_updated    DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(username)
);
```

### Migration File

`backend/database/migration/sqlite_migration_v5_v6.go` — adds the `UserSubscriptions` table.

### Go Model

`backend/models/subscription.go`:

```go
type UserSubscription struct {
    ID              int            `json:"id"`
    Username        string         `json:"username"`
    CreatorID       string         `json:"creator_id"`
    ImageTypes      SelectedTypes  `json:"image_types"`
    MediaScope      string         `json:"media_scope"`      // "all", "movies", "shows", "collections"
    LibrarySection  *string        `json:"library_section"`  // nil = all
    Enabled         bool           `json:"enabled"`
    DateCreated     time.Time      `json:"date_created"`
    DateUpdated     time.Time      `json:"date_updated"`
}
```

---

## 2. Backend Changes

### 2a. Database Layer

New file: `backend/database/sqlite_subscriptions.go`

| Function | Purpose |
|----------|---------|
| `CreateSubscription(ctx, sub)` | INSERT new subscription |
| `UpdateSubscription(ctx, sub)` | UPDATE existing subscription |
| `DeleteSubscription(ctx, id)` | DELETE subscription |
| `GetAllSubscriptions(ctx)` | SELECT all enabled subscriptions |
| `GetSubscriptionByUsername(ctx, username)` | Check if subscribed to a creator |
| `GetSubscriptionsByCreatorID(ctx, creatorID)` | Find all subscriptions for a creator (used by WebSocket handler) |

### 2b. API Routes

New file: `backend/routing/subscriptions/subscriptions.go`

| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| `GET` | `/api/subscriptions` | `GetAllSubscriptions` | List all subscriptions |
| `GET` | `/api/subscriptions/{username}` | `GetSubscription` | Get subscription for a specific creator |
| `POST` | `/api/subscriptions` | `CreateSubscription` | Create new subscription |
| `PUT` | `/api/subscriptions/{id}` | `UpdateSubscription` | Update subscription preferences |
| `DELETE` | `/api/subscriptions/{id}` | `DeleteSubscription` | Remove subscription |

Register in `backend/routing/routes.go` inside the protected route group.

### 2c. Subscription Matching Logic

New file: `backend/download/auto/subscription_check.go`

**`CheckSubscriptions(ctx)`** — called by existing cron job + WebSocket:

1. Fetch all enabled subscriptions from DB
2. For each subscription:
   a. Fetch creator's latest sets from MediUX API (reuse `mediux.GetAllUserSets()`)
   b. For each set, match against user's library items by TMDB ID
   c. For matched items not yet in `SavedItems`:
      - Create `DBSavedItem` with the subscription's `ImageTypes` and `AutoDownload=true`
      - Download images via existing download queue
      - Upsert to DB
3. Log results (new items added, items skipped, errors)

### 2d. WebSocket Integration

Modify `backend/download/auto/websocket-mediux.go`:

In the existing `update` event handler, after finding matching saved sets, also:

1. Check if any subscription exists for the set's `UserCreated.Username`
2. If yes and the set's TMDB ID matches a library item not yet in DB:
   - Auto-create `DBSavedItem` from subscription preferences
   - Trigger download via `CheckItem()`

### 2e. Cron Job

Modify `backend/jobs/download_auto.go`:

Add `subscriptionCheckInterval` (e.g., every 6 hours) that calls `CheckSubscriptions()`. This handles the case where the WebSocket misses events or the user wants to catch up on sets created while the app was offline.

### 2f. Notification Integration

Extend `backend/notification/` to support a new notification type:

- **New set from subscribed creator**: "New {type} set '{title}' by {username} applied to {item_title}"

Reuse existing notification providers (Discord, Pushover, Gotify, Webhook).

---

## 3. Frontend Changes

### 3a. Subscription Service

New file: `frontend/src/services/subscriptions/`

| File | Purpose |
|------|---------|
| `get-subscriptions.ts` | `GET /api/subscriptions` |
| `get-subscription.ts` | `GET /api/subscriptions/{username}` |
| `create-subscription.ts` | `POST /api/subscriptions` |
| `update-subscription.ts` | `PUT /api/subscriptions/{id}` |
| `delete-subscription.ts` | `DELETE /api/subscriptions/{id}` |

### 3b. Subscribe Button on User Page

Modify `frontend/src/app/user/[username]/page.tsx`:

- Add a "Subscribe" toggle button in the header area (next to username/avatar)
- When clicked, opens a `SubscriptionModal` to configure preferences
- Button shows active state (filled/highlighted) if already subscribed
- Button shows a gear icon if subscribed (to edit preferences)

### 3c. Subscription Modal

New file: `frontend/src/components/shared/subscription-modal.tsx`

Uses the same shadcn `Dialog` + `Form` pattern as `download-modal.tsx`:

```
┌─────────────────────────────────────────────┐
│  Subscribe to {username}                     │
│                                              │
│  Image Types:                                │
│  ☑ Posters  ☑ Backdrops  ☐ Season Posters   │
│  ☐ Special Season Posters  ☑ Titlecards      │
│                                              │
│  Download for:                               │
│  ◉ All  ○ Movies  ○ Shows  ○ Collections    │
│                                              │
│  Library Section:                            │
│  [All Sections ▾]                            │
│                                              │
│  ☑ Auto-download new sets                    │
│                                              │
│  [Cancel]  [Subscribe]                       │
└─────────────────────────────────────────────┘
```

### 3d. Subscriptions Management Page

New route: `frontend/src/app/subscriptions/page.tsx`

- Lists all active subscriptions with creator avatar, username, image types, scope
- Each subscription has: toggle enable/disable, edit preferences, unsubscribe
- Empty state: "Subscribe to creators to auto-download their sets"
- Link in Navbar or Settings to access

### 3e. Types

New file: `frontend/src/types/subscriptions/subscription.ts`:

```typescript
export interface UserSubscription {
  id: number;
  username: string;
  creator_id: string;
  image_types: SelectedTypes;
  media_scope: "all" | "movies" | "shows" | "collections";
  library_section: string | null;
  enabled: boolean;
  date_created: string;
  date_updated: string;
}
```

### 3f. Store

New file: `frontend/src/lib/stores/global-store-subscriptions.ts`:

Zustand store persisted to IndexedDB for subscription state.

---

## 4. Implementation Order

### Phase 1: Data Layer
1. Add migration `v5_v6` for `UserSubscriptions` table
2. Create `models/subscription.go`
3. Implement `database/sqlite_subscriptions.go` CRUD functions
4. Write unit tests for DB operations

### Phase 2: Backend API
5. Create `routing/subscriptions/` route handlers
6. Register routes in `routing/routes.go`
7. Test API endpoints with curl/Postman

### Phase 3: Auto-Download Integration
8. Implement `download/auto/subscription_check.go`
9. Modify WebSocket handler to check subscriptions
10. Add subscription check to cron job schedule
11. Add notification support for subscription downloads

### Phase 4: Frontend
12. Create subscription services (`services/subscriptions/`)
13. Create subscription types and store
14. Build `SubscriptionModal` component
15. Add subscribe button to user page
16. Build subscriptions management page
17. Add navigation link

### Phase 5: Polish
18. Add loading states, error handling, toast notifications
19. Test edge cases (already saved sets, duplicate detection, library matching)
20. Update Swagger docs (`generate_go_docs.sh`)

---

## 5. Edge Cases & Considerations

| Case | Handling |
|------|----------|
| **Set already in DB** | Skip — don't re-download if `SavedItems` already has this `(tmdb_id, poster_set_id)` |
| **Multiple subscriptions overlap** | Last write wins on `SelectedTypes` per item (existing behavior in `UpsertSavedItem`) |
| **Creator deletes a set** | No action needed — existing set remains, auto-download just stops finding it |
| **Library section mismatch** | Only match sets whose TMDB IDs exist in items from the user's selected library section |
| **Subscription while offline** | Cron job catches up on next run — fetches all creator sets and matches against library |
| **Unsubscribe** | Keep existing saved sets, just stop auto-downloading new ones |
| **Wildcard types** | Subscribe with all types selected = download everything from that creator |
| **Collection sets** | Handle `auto_add_new_collection_items` — when a collection gains new members, download their images too |

---

## 6. Files to Create/Modify

### New Files
```
backend/models/subscription.go
backend/database/sqlite_subscriptions.go
backend/database/migration/sqlite_migration_v5_v6.go
backend/routing/subscriptions/subscriptions.go
backend/download/auto/subscription_check.go
frontend/src/services/subscriptions/get-subscriptions.ts
frontend/src/services/subscriptions/get-subscription.ts
frontend/src/services/subscriptions/create-subscription.ts
frontend/src/services/subscriptions/update-subscription.ts
frontend/src/services/subscriptions/delete-subscription.ts
frontend/src/types/subscriptions/subscription.ts
frontend/src/lib/stores/global-store-subscriptions.ts
frontend/src/components/shared/subscription-modal.tsx
frontend/src/app/subscriptions/page.tsx
```

### Modified Files
```
backend/routing/routes.go                          — register subscription routes
backend/download/auto/websocket-mediux.go          — check subscriptions on update events
backend/jobs/download_auto.go                      — add subscription check cron
backend/database/migration/migrate.go              — add v5→v6 migration
frontend/src/app/user/[username]/page.tsx           — add subscribe button
frontend/src/components/layout/app-navbar.tsx       — add subscriptions link
```
