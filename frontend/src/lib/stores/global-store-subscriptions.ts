import { create } from "zustand";
import { persist } from "zustand/middleware";

import { GlobalStore } from "@/lib/stores/stores";

import type { UserSubscription } from "@/types/subscriptions/subscription";

interface SubscriptionStore {
  subscriptions: UserSubscription[];
  setSubscriptions: (subs: UserSubscription[]) => void;
  addSubscription: (sub: UserSubscription) => void;
  removeSubscription: (id: number) => void;
  updateSubscription: (sub: UserSubscription) => void;
  getSubscriptionByUsername: (username: string) => UserSubscription | undefined;

  hasHydrated: boolean;
  hydrate: () => void;
  clear: () => void;
}

export const useSubscriptionStore = create<SubscriptionStore>()(
  persist(
    (set, get) => ({
      subscriptions: [],
      setSubscriptions: (subs) => set({ subscriptions: subs }),
      addSubscription: (sub) => set((state) => ({ subscriptions: [...state.subscriptions, sub] })),
      removeSubscription: (id) => set((state) => ({ subscriptions: state.subscriptions.filter((s) => s.id !== id) })),
      updateSubscription: (sub) =>
        set((state) => ({
          subscriptions: state.subscriptions.map((s) => (s.id === sub.id ? sub : s)),
        })),
      getSubscriptionByUsername: (username) => get().subscriptions.find((s) => s.username === username),

      hasHydrated: false,
      hydrate: () => set({ hasHydrated: true }),
      clear: () => set({ subscriptions: [] }),
    }),
    {
      name: "Subscriptions",
      storage: GlobalStore,
      partialize: (state) => ({
        subscriptions: state.subscriptions,
      }),
      onRehydrateStorage: () => (state) => {
        state?.hydrate();
      },
    }
  )
);
