import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { UserSubscription } from "@/types/subscriptions/subscription";

export interface GetAllSubscriptions_Response {
  subscriptions: UserSubscription[];
}

export const GetAllSubscriptions = async (): Promise<APIResponse<GetAllSubscriptions_Response>> => {
  log("INFO", "API - Subscriptions", "Fetch All Subscriptions", "Fetching all subscriptions");
  try {
    const response = await apiClient.get<APIResponse<GetAllSubscriptions_Response>>("/subscriptions");
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || "Unknown error fetching subscriptions");
    } else {
      log("INFO", "API - Subscriptions", "Fetch All Subscriptions", "Fetched all subscriptions successfully", response.data);
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Subscriptions", "Fetch All Subscriptions", `Failed to fetch subscriptions: ${error instanceof Error ? error.message : "Unknown error"}`, error);
    return ReturnErrorMessage<GetAllSubscriptions_Response>(error);
  }
};
