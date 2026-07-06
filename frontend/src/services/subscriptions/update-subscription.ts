import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";
import type { UpdateSubscriptionRequest } from "@/types/subscriptions/subscription";

export interface UpdateSubscription_Response {
  success: boolean;
}

export const UpdateSubscription = async (id: number, request: UpdateSubscriptionRequest): Promise<APIResponse<UpdateSubscription_Response>> => {
  log("INFO", "API - Subscriptions", "Update Subscription", `Updating subscription ${id}`);
  try {
    const response = await apiClient.put<APIResponse<UpdateSubscription_Response>>(`/subscriptions/${id}`, request);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || `Unknown error updating subscription ${id}`);
    } else {
      log("INFO", "API - Subscriptions", "Update Subscription", `Updated subscription ${id} successfully`, response.data);
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Subscriptions", "Update Subscription", `Failed to update subscription ${id}: ${error instanceof Error ? error.message : "Unknown error"}`, error);
    return ReturnErrorMessage<UpdateSubscription_Response>(error);
  }
};
