import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { UserSubscription } from "@/types/subscriptions/subscription";

export interface GetSubscriptionByUsername_Response {
  subscription: UserSubscription | null;
  subscribed: boolean;
}

export const GetSubscriptionByUsername = async (username: string): Promise<APIResponse<GetSubscriptionByUsername_Response>> => {
  log("INFO", "API - Subscriptions", "Fetch Subscription by Username", `Fetching subscription for ${username}`);
  try {
    const params = { username };
    const response = await apiClient.get<APIResponse<GetSubscriptionByUsername_Response>>("/subscriptions/username", { params });
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || `Unknown error fetching subscription for ${username}`);
    } else {
      log("INFO", "API - Subscriptions", "Fetch Subscription by Username", `Fetched subscription for ${username} successfully`, response.data);
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Subscriptions", "Fetch Subscription by Username", `Failed to fetch subscription for ${username}: ${error instanceof Error ? error.message : "Unknown error"}`, error);
    return ReturnErrorMessage<GetSubscriptionByUsername_Response>(error);
  }
};
