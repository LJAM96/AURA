"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { Bell, BellOff, Loader } from "lucide-react";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { toast } from "sonner";
import { z } from "zod";

import { Button } from "@/components/ui/button";
import { Checkbox } from "@/components/ui/checkbox";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel } from "@/components/ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";

import { cn } from "@/lib/cn";
import { useSubscriptionStore } from "@/lib/stores/global-store-subscriptions";
import { CreateSubscription } from "@/services/subscriptions/create-subscription";
import { DeleteSubscription } from "@/services/subscriptions/delete-subscription";
import { UpdateSubscription } from "@/services/subscriptions/update-subscription";

import type { SelectedTypes } from "@/types/media-and-posters/media-item-and-library";
import { DOWNLOAD_IMAGE_TYPE_OPTIONS, TYPE_DOWNLOAD_IMAGE_TYPE_OPTIONS } from "@/types/ui-options";
import type { CreateSubscriptionRequest, UserSubscription, TYPE_PRIORITY_OPTIONS } from "@/types/subscriptions/subscription";
import { PRIORITY_OPTIONS } from "@/types/subscriptions/subscription";

interface SubscriptionModalProps {
  username: string;
  creatorId?: string;
  existingSubscription?: UserSubscription;
  triggerClassName?: string;
  triggerChildren?: React.ReactNode;
}

const formSchema = z.object({
  types: z.array(z.enum(["poster", "backdrop", "season_poster", "special_season_poster", "titlecard"])).min(1, {
    message: "Select at least one image type.",
  }),
  media_scope: z.enum(["all", "movies", "shows", "collections"]),
  priority: z.number().min(1).max(5),
});

type FormValues = z.infer<typeof formSchema>;

export function SubscriptionModal({ username, creatorId, existingSubscription, triggerClassName, triggerChildren }: SubscriptionModalProps) {
  const [open, setOpen] = useState(false);
  const [loading, setLoading] = useState(false);
  const { addSubscription, removeSubscription, updateSubscription } = useSubscriptionStore();

  const isSubscribed = !!existingSubscription;

  const form = useForm<FormValues>({
    resolver: zodResolver(formSchema),
    defaultValues: {
      types: isSubscribed
        ? [
            ...(existingSubscription.image_types.poster ? ["poster"] : []),
            ...(existingSubscription.image_types.backdrop ? ["backdrop"] : []),
            ...(existingSubscription.image_types.season_poster ? ["season_poster"] : []),
            ...(existingSubscription.image_types.special_season_poster ? ["special_season_poster"] : []),
            ...(existingSubscription.image_types.titlecard ? ["titlecard"] : []),
          ] as TYPE_DOWNLOAD_IMAGE_TYPE_OPTIONS[]
        : ["poster", "titlecard"],
      media_scope: existingSubscription?.media_scope || "all",
      priority: existingSubscription?.priority || 1,
    },
  });

  const onSubmit = async (data: FormValues) => {
    setLoading(true);
    try {
      const imageTypes: SelectedTypes = {
        poster: data.types.includes("poster"),
        backdrop: data.types.includes("backdrop"),
        season_poster: data.types.includes("season_poster"),
        special_season_poster: data.types.includes("special_season_poster"),
        titlecard: data.types.includes("titlecard"),
      };

      if (isSubscribed && existingSubscription) {
        const response = await UpdateSubscription(existingSubscription.id, {
          creator_id: creatorId || "",
          image_types: imageTypes,
          priority: data.priority as TYPE_PRIORITY_OPTIONS,
          media_scope: data.media_scope,
          library_section: null,
          enabled: true,
        });
        if (response.data?.success) {
          updateSubscription({
            ...existingSubscription,
            image_types: imageTypes,
            priority: data.priority as TYPE_PRIORITY_OPTIONS,
            media_scope: data.media_scope,
          });
          toast.success(`Updated subscription for ${username}`);
          setOpen(false);
        } else {
          toast.error(response.error?.message || "Failed to update subscription");
        }
      } else {
        const request: CreateSubscriptionRequest = {
          username,
          creator_id: creatorId || "",
          image_types: imageTypes,
          priority: data.priority as TYPE_PRIORITY_OPTIONS,
          media_scope: data.media_scope,
          library_section: null,
          enabled: true,
        };
        const response = await CreateSubscription(request);
        if (response.data?.id) {
          addSubscription({
            id: response.data.id,
            username,
            creator_id: creatorId || "",
            image_types: imageTypes,
            priority: data.priority as TYPE_PRIORITY_OPTIONS,
            media_scope: data.media_scope,
            library_section: null,
            enabled: true,
            date_created: new Date().toISOString(),
            date_updated: new Date().toISOString(),
          });
          toast.success(`Subscribed to ${username}`);
          setOpen(false);
        } else {
          toast.error(response.error?.message || "Failed to create subscription");
        }
      }
    } catch {
      toast.error("An unexpected error occurred");
    } finally {
      setLoading(false);
    }
  };

  const handleUnsubscribe = async () => {
    if (!existingSubscription) return;
    setLoading(true);
    try {
      const response = await DeleteSubscription(existingSubscription.id);
      if (response.data?.success) {
        removeSubscription(existingSubscription.id);
        toast.success(`Unsubscribed from ${username}`);
        setOpen(false);
      } else {
        toast.error(response.error?.message || "Failed to delete subscription");
      }
    } catch {
      toast.error("An unexpected error occurred");
    } finally {
      setLoading(false);
    }
  };

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        setOpen(v);
        if (v) {
          form.reset();
        }
      }}
    >
      <DialogTrigger asChild>
        {triggerChildren || (
          <Button variant={isSubscribed ? "default" : "outline"} size="sm" className={cn("gap-2", triggerClassName)}>
            {isSubscribed ? <Bell className="h-4 w-4" /> : <BellOff className="h-4 w-4" />}
            {isSubscribed ? "Subscribed" : "Subscribe"}
          </Button>
        )}
      </DialogTrigger>
      <DialogPortal>
        <DialogOverlay />
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{isSubscribed ? `Edit Subscription to ${username}` : `Subscribe to ${username}`}</DialogTitle>
            <DialogDescription>
              {isSubscribed
                ? "Update which image types to auto-download from this creator."
                : "Automatically download new sets from this creator matching your selected image types."}
            </DialogDescription>
          </DialogHeader>

          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="types"
                render={() => (
                  <FormItem>
                    <div className="mb-4">
                      <FormLabel className="text-base">Image Types</FormLabel>
                    </div>
                    <div className="grid grid-cols-2 gap-2">
                      {DOWNLOAD_IMAGE_TYPE_OPTIONS.map((option) => (
                        <FormField
                          key={option.value}
                          control={form.control}
                          name="types"
                          render={({ field }) => (
                            <FormItem key={option.value} className="flex flex-row items-start space-x-3 space-y-0">
                              <FormControl>
                                <Checkbox
                                  checked={field.value?.includes(option.value)}
                                  onCheckedChange={(checked) => {
                                    return checked ? field.onChange([...field.value, option.value]) : field.onChange(field.value?.filter((value) => value !== option.value));
                                  }}
                                />
                              </FormControl>
                              <FormLabel className="font-normal">{option.label}</FormLabel>
                            </FormItem>
                          )}
                        />
                      ))}
                    </div>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="media_scope"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Download for</FormLabel>
                    <Select onValueChange={field.onChange} defaultValue={field.value}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select media scope" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="all">All</SelectItem>
                        <SelectItem value="movies">Movies</SelectItem>
                        <SelectItem value="shows">Shows</SelectItem>
                        <SelectItem value="collections">Collections</SelectItem>
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />

              <FormField
                control={form.control}
                name="priority"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Priority</FormLabel>
                    <FormDescription className="text-xs text-muted-foreground">
                      When multiple subscriptions claim the same image type for an item, higher priority (lower number) wins. If the higher priority creator doesn't have the image, lower priority creators are used as fallback.
                    </FormDescription>
                    <Select onValueChange={(value) => field.onChange(parseInt(value))} defaultValue={field.value?.toString()}>
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select priority" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {PRIORITY_OPTIONS.map((option) => (
                          <SelectItem key={option.value} value={option.value.toString()}>
                            {option.label}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                  </FormItem>
                )}
              />

              <DialogFooter className="flex justify-between">
                {isSubscribed ? (
                  <Button type="button" variant="destructive" onClick={handleUnsubscribe} disabled={loading}>
                    Unsubscribe
                  </Button>
                ) : (
                  <DialogClose asChild>
                    <Button type="button" variant="outline">
                      Cancel
                    </Button>
                  </DialogClose>
                )}
                <Button type="submit" disabled={loading}>
                  {loading && <Loader className="mr-2 h-4 w-4 animate-spin" />}
                  {isSubscribed ? "Update" : "Subscribe"}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  );
}
