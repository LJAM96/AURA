"use client";

import { GetAllSubscriptions } from "@/services/subscriptions/get-subscriptions";
import { formatLastUpdatedDate } from "@/helper/format-date-last-updates";
import Loader from "@/components/shared/loader";
import { Bell, BellOff, Settings, Trash2 } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { toast } from "sonner";

import { SubscriptionModal } from "@/components/shared/subscription-modal";
import { Avatar, AvatarFallback, AvatarImage } from "@/components/ui/avatar";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { DeleteSubscription } from "@/services/subscriptions/delete-subscription";

import { useSubscriptionStore } from "@/lib/stores/global-store-subscriptions";

import type { UserSubscription } from "@/types/subscriptions/subscription";
import { MEDIA_SCOPE_OPTIONS } from "@/types/subscriptions/subscription";

export default function SubscriptionsPage() {
  const { subscriptions, setSubscriptions, removeSubscription } = useSubscriptionStore();
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchSubscriptions = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const response = await GetAllSubscriptions();
      if (response.data?.subscriptions) {
        setSubscriptions(response.data.subscriptions);
      } else if (response.error) {
        setError(response.error.message || "Failed to load subscriptions");
      }
    } catch {
      setError("An unexpected error occurred");
    } finally {
      setLoading(false);
    }
  }, [setSubscriptions]);

  useEffect(() => {
    fetchSubscriptions();
  }, [fetchSubscriptions]);

  const handleDelete = async (sub: UserSubscription) => {
    try {
      const response = await DeleteSubscription(sub.id);
      if (response.data?.success) {
        removeSubscription(sub.id);
        toast.success(`Unsubscribed from ${sub.username}`);
      } else {
        toast.error(response.error?.message || "Failed to delete subscription");
      }
    } catch {
      toast.error("An unexpected error occurred");
    }
  };

  const getImageTypeLabels = (sub: UserSubscription) => {
    const labels: string[] = [];
    if (sub.image_types.poster) labels.push("Poster");
    if (sub.image_types.backdrop) labels.push("Backdrop");
    if (sub.image_types.season_poster) labels.push("Season Poster");
    if (sub.image_types.special_season_poster) labels.push("Special Season Poster");
    if (sub.image_types.titlecard) labels.push("Titlecard");
    return labels;
  };

  const getScopeLabel = (scope: string) => {
    return MEDIA_SCOPE_OPTIONS.find((opt) => opt.value === scope)?.label || scope;
  };

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <div>
          <h1 className="text-3xl font-bold">Subscriptions</h1>
          <p className="text-muted-foreground mt-1">Manage your creator subscriptions for automatic downloads.</p>
        </div>
      </div>

      {loading && (
        <div className="flex justify-center py-12">
          <Loader message="Loading subscriptions..." />
        </div>
      )}

      {error && (
        <div className="text-center py-12">
          <p className="text-destructive">{error}</p>
          <Button variant="outline" onClick={fetchSubscriptions} className="mt-4">
            Retry
          </Button>
        </div>
      )}

      {!loading && !error && subscriptions.length === 0 && (
        <div className="text-center py-12">
          <BellOff className="h-12 w-12 mx-auto text-muted-foreground mb-4" />
          <h2 className="text-xl font-semibold mb-2">No Subscriptions</h2>
          <p className="text-muted-foreground">
            Visit a creator&apos;s page and click Subscribe to automatically download their new sets.
          </p>
        </div>
      )}

      {!loading && !error && subscriptions.length > 0 && (
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {subscriptions.map((sub) => (
            <Card key={sub.id}>
              <CardHeader className="pb-3">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-3">
                    <Avatar className="h-10 w-10">
                      <AvatarImage src={`/api/images/mediux/avatar?username=${sub.username}`} />
                      <AvatarFallback>{sub.username.charAt(0).toUpperCase()}</AvatarFallback>
                    </Avatar>
                    <div>
                      <CardTitle className="text-lg">{sub.username}</CardTitle>
                      <CardDescription>
                        {sub.enabled ? (
                          <Badge variant="default" className="text-xs">
                            <Bell className="h-3 w-3 mr-1" />
                            Active
                          </Badge>
                        ) : (
                          <Badge variant="secondary" className="text-xs">
                            <BellOff className="h-3 w-3 mr-1" />
                            Paused
                          </Badge>
                        )}
                      </CardDescription>
                    </div>
                  </div>
                  <div className="flex gap-1">
                    <SubscriptionModal
                      username={sub.username}
                      existingSubscription={sub}
                      triggerChildren={
                        <Button variant="ghost" size="icon" className="h-8 w-8">
                          <Settings className="h-4 w-4" />
                        </Button>
                      }
                    />
                    <Button variant="ghost" size="icon" className="h-8 w-8 text-destructive" onClick={() => handleDelete(sub)}>
                      <Trash2 className="h-4 w-4" />
                    </Button>
                  </div>
                </div>
              </CardHeader>
              <CardContent>
                <div className="space-y-2 text-sm">
                  <div>
                    <span className="text-muted-foreground">Scope:</span>{" "}
                    <span className="font-medium">{getScopeLabel(sub.media_scope)}</span>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Priority:</span>{" "}
                    <Badge variant="secondary" className="text-xs">
                      {sub.priority === 1 ? "1 (Highest)" : sub.priority === 5 ? "5 (Lowest)" : sub.priority}
                    </Badge>
                  </div>
                  <div>
                    <span className="text-muted-foreground">Image Types:</span>
                    <div className="flex flex-wrap gap-1 mt-1">
                      {getImageTypeLabels(sub).map((label) => (
                        <Badge key={label} variant="outline" className="text-xs">
                          {label}
                        </Badge>
                      ))}
                    </div>
                  </div>
                  <div className="text-muted-foreground text-xs pt-2">
                    Subscribed {formatLastUpdatedDate(sub.date_created, sub.date_created)}
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}
    </div>
  );
}
