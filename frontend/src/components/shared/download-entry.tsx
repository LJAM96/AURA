"use client";

import { formatExactDateTime } from "@/helper/format-date-last-updates";
import { formatImageTypeLabel } from "@/helper/format-image-type";
import { RemoveDownloadHistoryEntry } from "@/services/downloads/history-remove";
import { RemoveItemFromQueue } from "@/services/downloads/queue-remove";
import { Trash2 } from "lucide-react";
import { toast } from "sonner";

import React from "react";

import Link from "next/link";

import { AssetImage } from "@/components/shared/asset-image";
import { ConfirmDestructiveDialogActionButton } from "@/components/shared/dialog-destructive-action";
import { renderTypeBadges } from "@/components/shared/saved-sets-shared";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";
import { Badge } from "@/components/ui/badge";
import { Card, CardContent, CardHeader } from "@/components/ui/card";
import { Separator } from "@/components/ui/separator";
import { H4 } from "@/components/ui/typography";

import { cn } from "@/lib/cn";
import { useMediaStore } from "@/lib/stores/global-store-media-store";

import type { DownloadHistoryEntry, DownloadQueueJob } from "@/types/database/download-queue";
import type { MediaItem } from "@/types/media-and-posters/media-item-and-library";

const statusStyles: Record<DownloadHistoryEntry["status"], string> = {
  success: "border-green-400 text-green-500",
  warning: "border-yellow-400 text-yellow-500",
  error: "border-red-400 text-red-500",
};

type DownloadEntryCardProps =
  | {
      mode: "queue";
      job: DownloadQueueJob;
      fetchQueueEntries?: () => Promise<void>;
    }
  | {
      mode: "history";
      entry: DownloadHistoryEntry;
      fetchHistory?: () => Promise<void>;
    };

const DownloadEntryCard: React.FC<DownloadEntryCardProps> = (props) => {
  const { setMediaItem } = useMediaStore();

  const isQueue = props.mode === "queue";

  const mediaItem: MediaItem | undefined = isQueue ? props.job.item.media_item : undefined;
  const posterSets = isQueue ? (Array.isArray(props.job.item.poster_sets) ? props.job.item.poster_sets : []) : [];

  const title = isQueue ? mediaItem!.title : props.entry.media_item_title;
  const year = isQueue ? mediaItem!.year : props.entry.media_item_year;
  const libraryTitle = isQueue ? mediaItem!.library_title : props.entry.library_title;
  const edition = isQueue ? mediaItem!.edition : props.entry.edition;
  const ratingKey = isQueue ? mediaItem!.rating_key : props.entry.rating_key;

  const onDeleteConfirm = async () => {
    if (isQueue) {
      try {
        const response = await RemoveItemFromQueue(props.job.id);
        if (response.status === "error") {
          toast.error(
            `Error deleting from queue: ${response.error?.message || "Unknown error occurred trying to delete."}`
          );
        } else {
          toast.success(response.data?.result || "Successfully deleted from queue.");
        }
      } catch (error) {
        toast.error(
          `Error deleting from queue: ${error instanceof Error ? error.message : "Unknown error occurred trying to delete."}`
        );
      }
      if (props.fetchQueueEntries) {
        await props.fetchQueueEntries();
      }
    } else {
      try {
        const response = await RemoveDownloadHistoryEntry(props.entry.id);
        if (response.status === "error") {
          toast.error(`Error deleting history entry: ${response.error?.message || "Unknown error occurred."}`);
        } else {
          toast.success(response.data?.result || "Successfully deleted history entry.");
        }
      } catch (error) {
        toast.error(
          `Error deleting history entry: ${error instanceof Error ? error.message : "Unknown error occurred."}`
        );
      }
      if (props.fetchHistory) {
        await props.fetchHistory();
      }
    }
  };

  return (
    <Card className="relative w-full max-w-md mx-auto">
      <CardHeader>
        <div className="absolute top-2 left-2">
          <ConfirmDestructiveDialogActionButton
            variant="outline"
            className="text-destructive border-1 shadow-none hover:text-red-500 cursor-pointer"
            confirmText={isQueue ? "Delete File" : "Delete Entry"}
            title={isQueue ? "Delete Downloaded File?" : "Delete History Entry?"}
            description={
              isQueue
                ? "Are you sure you want to delete the downloaded file for this media item? This action cannot be undone."
                : "Are you sure you want to delete this download history entry? This action cannot be undone."
            }
            onConfirm={onDeleteConfirm}
          >
            <Trash2 className="w-5 h-5" />
          </ConfirmDestructiveDialogActionButton>
        </div>
        <div className="absolute top-2 right-2">
          {!isQueue && (
            <Badge variant="outline" className={cn("text-sm", statusStyles[props.entry.status])}>
              {props.entry.status.toUpperCase()}
            </Badge>
          )}
        </div>
      </CardHeader>

      {isQueue ? (
        <div className="flex justify-center">
          <AssetImage
            image={mediaItem!}
            imageType="item"
            className="w-[80%] h-auto transition-transform hover:scale-105"
          />
        </div>
      ) : (
        ratingKey && (
          <div className="flex justify-center">
            <AssetImage image={{ rating_key: ratingKey } as MediaItem} imageType="item" className="w-[80%] h-auto" />
          </div>
        )
      )}

      <CardContent className={cn("p-0 ml-2 mr-2", isQueue ? "" : "mt-6")}>
        <H4>
          {isQueue ? (
            <Link
              href={"/media-item/"}
              className="text-primary hover:underline"
              onClick={() => {
                setMediaItem(mediaItem!);
              }}
            >
              {title}
            </Link>
          ) : (
            title
          )}
        </H4>

        <span className="text-xs sm:text-sm text-muted-foreground inline-block">
          {year} · {libraryTitle}
          {edition ? ` · ${edition}` : ""}
        </span>

        {isQueue ? (
          posterSets.length > 0 && (
            <span className="text-xs sm:text-sm text-muted-foreground inline-block">
              {`Set by: ${posterSets[0].user_created}`}
            </span>
          )
        ) : (
          <>
            <span className="text-xs sm:text-sm text-muted-foreground inline-block">{`Set: ${props.entry.set_title}`}</span>
            <span className="text-xs text-muted-foreground inline-block">
              {formatExactDateTime(props.entry.created_at)}
            </span>
          </>
        )}

        <Separator className="my-4" />

        {isQueue ? (
          posterSets.some(
            (set) =>
              set.selected_types.poster ||
              set.selected_types.backdrop ||
              set.selected_types.season_poster ||
              set.selected_types.titlecard
          ) ? (
            <div className="flex flex-wrap gap-2">{renderTypeBadges(props.job.item)}</div>
          ) : (
            <div className="flex flex-wrap gap-2">
              <Badge key={"no-types"} variant="outline" className="text-sm bg-red-500">
                No Selected Types
              </Badge>
            </div>
          )
        ) : (
          <>
            <div className="flex flex-wrap gap-2">
              <Badge variant="outline" className="text-sm border-green-400 text-green-500">
                {props.entry.images_succeeded} Succeeded
              </Badge>
              {props.entry.images_failed > 0 && (
                <Badge variant="outline" className="text-sm border-red-400 text-red-500">
                  {props.entry.images_failed} Failed
                </Badge>
              )}
            </div>

            {props.entry.failed_images.length > 0 && (
              <Accordion type="single" collapsible className="mt-3">
                <AccordionItem value="failed_images">
                  <AccordionTrigger className="cursor-pointer text-sm">Failed Images</AccordionTrigger>
                  <AccordionContent>
                    <ul className="text-xs space-y-1">
                      {props.entry.failed_images.map((image, idx) => {
                        const label = formatImageTypeLabel(image);
                        return (
                          <li key={idx} className="text-red-500">
                            {label && <b>{label}</b>}
                            {image.failure_reason ? `${label ? ": " : ""}${image.failure_reason}` : ""}
                          </li>
                        );
                      })}
                    </ul>
                  </AccordionContent>
                </AccordionItem>
              </Accordion>
            )}
          </>
        )}
      </CardContent>
    </Card>
  );
};

export default DownloadEntryCard;
