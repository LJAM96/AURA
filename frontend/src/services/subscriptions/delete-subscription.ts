import apiClient from "@/services/api-client";
import { ReturnErrorMessage } from "@/services/api-error-return";

import { log } from "@/lib/logger";

import type { APIResponse } from "@/types/api/api-response";

export interface DeleteSubscription_Response {
  success: boolean;
}

export const DeleteSubscription = async (id: number): Promise<APIResponse<DeleteSubscription_Response>> => {
  log("INFO", "API - Subscriptions", "Delete Subscription", `Deleting subscription ${id}`);
  try {
    const response = await apiClient.delete<APIResponse<DeleteSubscription_Response>>(`/subscriptions/${id}`);
    if (response.data.status === "error") {
      throw new Error(response.data.error?.message || `Unknown error deleting subscription ${id}`);
    } else {
      log("INFO", "API - Subscriptions", "Delete Subscription", `Deleted subscription ${id} successfully`, response.data);
    }
    return response.data;
  } catch (error) {
    log("ERROR", "API - Subscriptions", "Delete Subscription", `Failed to delete subscription ${id}: ${error instanceof Error ? error.message : "Unknown error"}`, error);
    return ReturnErrorMessage<DeleteSubscription_Response>(error);
  }
};
