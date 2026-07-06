import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { CreateSubscriptionRequest } from "@/types/subscriptions/subscription";

export interface CreateSubscription_Response {
  id: number;
}

export const CreateSubscription = async (request: CreateSubscriptionRequest): Promise<APIResponse<CreateSubscription_Response>> => {
  log("INFO", "API - Subscriptions", "Create Subscription", `Creating subscription for ${request.username}`);
  try {
    const response = await apiClient.post<APIResponse<CreateSubscription_Response>>("/subscriptions", request);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || `Unknown error creating subscription for ${request.username}`);
    } else {
      log("INFO", "API - Subscriptions", "Create Subscription", `Created subscription for ${request.username} successfully`, response.data);
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Subscriptions", "Create Subscription", `Failed to create subscription for ${request.username}: ${error instanceof Error ? error.message : "Unknown error"}`, error);
    return ReturnErrorMessage<CreateSubscription_Response>(error);
  }
};
