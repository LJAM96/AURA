"use client";

import { Suspense, useEffect, useMemo, useState } from "react";

import { useSearchParams } from "next/navigation";

import { ChangelogMarkdown } from "@/components/shared/changelog-markdown";
import { Accordion, AccordionContent, AccordionItem, AccordionTrigger } from "@/components/ui/accordion";

export default function Changelog() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <ChangelogContent />
    </Suspense>
  );
}

interface ChangelogSection {
  version: string | null;
  date: string | null;
  content: string;
}

function splitChangelogSections(markdown: string): ChangelogSection[] {
  if (!markdown) return [];
  return markdown
    .split(/(?=^##\s*\[)/m)
    .filter((block) => block.trim().length > 0)
    .map((block) => {
      const trimmed = block.replace(/\n-{3,}\s*$/, "").trim();
      const match = /##\s*\[([^\]]+)\]\s*-\s*(\d{4}-\d{2}-\d{2})/.exec(trimmed);
      return {
        version: match ? match[1] : null,
        date: match ? match[2] : null,
        content: trimmed,
      };
    });
}

const normalizeVersion = (v: string) => v.replace(/^v/, "").replace(/-.*$/, "");

function ChangelogContent() {
  const [content, setContent] = useState("");
  const searchParams = useSearchParams();
  const [currentVersion, setCurrentVersion] = useState<string | null>(null);
  const [latestVersion, setLatestVersion] = useState<string | null>(null);

  useEffect(() => {
    const currentVersion = searchParams.get("currentVersion");
    setCurrentVersion(currentVersion);
    const updates = searchParams.get("updates");
    const latestVersion = searchParams.get("latestVersion");
    setLatestVersion(latestVersion);
    if (updates === "true") {
      // Fetch latest changelog from GitHub (raw URL)
      fetch("https://raw.githubusercontent.com/mediux-team/AURA/master/frontend/public/CHANGELOG.md")
        .then((res) => res.text())
        .then(setContent)
        .catch(() => setContent("Failed to fetch latest changelog from GitHub."));
    } else {
      const currentVersion = searchParams.get("currentVersion");
      setCurrentVersion(currentVersion);
      const latestVersion = searchParams.get("latestVersion");
      setLatestVersion(latestVersion);
      // Fetch local changelog
      fetch("/CHANGELOG.md")
        .then((res) => res.text())
        .then(setContent)
        .catch(() => setContent("Failed to fetch local changelog."));
    }
  }, [searchParams]);

  const sections = useMemo(() => splitChangelogSections(content), [content]);
  const [latestSection, ...olderSections] = sections;

  // Auto-expand the section matching currentVersion if it isn't the latest one already shown open.
  const currentSectionValue = useMemo(() => {
    if (!currentVersion) return undefined;
    const idx = olderSections.findIndex(
      (section) => section.version && normalizeVersion(section.version) === normalizeVersion(currentVersion)
    );
    return idx >= 0 ? `version-${idx}` : undefined;
  }, [olderSections, currentVersion]);

  return (
    <div className="min-h-screen py-2 px-10 flex justify-center">
      <div className="w-full max-w-4xl rounded-lg shadow-md p-2">
        <h1 className="text-3xl font-bold mb-6 text-center">Change Log</h1>

        {latestVersion && currentVersion && (
          <>
            <div className="my-6 flex items-center gap-3">
              <hr className="flex-grow border-amber-400 border-t-2" />
            </div>
            <h2 className="text-xl font-bold mb-4 text-amber-700 text-center">Updates since {currentVersion}</h2>
          </>
        )}

        {latestSection && (
          <ChangelogMarkdown currentVersion={currentVersion}>{latestSection.content}</ChangelogMarkdown>
        )}

        {olderSections.length > 0 && (
          <>
            <div className="my-8 flex items-center gap-3">
              <hr className="flex-grow border-muted-foreground/20" />
              <span className="text-muted-foreground text-sm font-semibold uppercase tracking-wide">
                Previous Updates
              </span>
              <hr className="flex-grow border-muted-foreground/20" />
            </div>
            <Accordion type="multiple" defaultValue={currentSectionValue ? [currentSectionValue] : []}>
              {olderSections.map((section, index) => (
                <AccordionItem key={section.version ?? index} value={`version-${index}`}>
                  <AccordionTrigger className="text-lg font-bold">
                    <span className="flex items-center gap-3">
                      {section.version ?? "Unknown version"}
                      {section.date && (
                        <span className="text-muted-foreground text-sm font-normal">{section.date}</span>
                      )}
                      {`version-${index}` === currentSectionValue && (
                        <span className="bg-yellow-100 text-yellow-800 px-2 py-0.5 rounded text-xs font-semibold border border-yellow-300">
                          You are here
                        </span>
                      )}
                    </span>
                  </AccordionTrigger>
                  <AccordionContent>
                    <ChangelogMarkdown currentVersion={currentVersion}>
                      {section.content.replace(/^##\s*\[[^\]]+\]\s*-\s*\d{4}-\d{2}-\d{2}\s*/, "")}
                    </ChangelogMarkdown>
                  </AccordionContent>
                </AccordionItem>
              ))}
            </Accordion>
          </>
        )}
      </div>
    </div>
  );
}
