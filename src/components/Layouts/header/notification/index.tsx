"use client";

import {
  Dropdown,
  DropdownContent,
  DropdownTrigger,
} from "@/components/ui/dropdown";
import { useIsMobile } from "@/hooks/use-mobile";
import { cn } from "@/lib/utils";
import Image from "next/image";
import Link from "next/link";
import { useState, useEffect } from "react";
import { BellIcon } from "./icons";
import dayjs from "dayjs";

type ActivityLog = {
  id: bigint;
  action: string;
  details: string | null;
  createdAt: Date;
  user: {
    id: number;
    username: string;
  };
};

export function Notification() {
  const [isOpen, setIsOpen] = useState(false);
  const [isDotVisible, setIsDotVisible] = useState(true);
  const [logs, setLogs] = useState<ActivityLog[]>([]);
  const [loading, setLoading] = useState(true);
  const isMobile = useIsMobile();

  useEffect(() => {
    if (isOpen) {
      fetchLogs();
    }
  }, [isOpen]);

  useEffect(() => {
    // Poll for new logs every 30 seconds
    const interval = setInterval(() => {
      if (!isOpen) {
        fetchLogs();
      }
    }, 30000);

    return () => clearInterval(interval);
  }, [isOpen]);

  const fetchLogs = async () => {
    try {
      const res = await fetch("/api/activity-logs");
      if (res.ok) {
        const data = await res.json();
        setLogs(data);
        setLoading(false);
      }
    } catch (error) {
      console.error("Failed to fetch activity logs", error);
      setLoading(false);
    }
  };

  const getActionText = (action: string, details?: string | null) => {
    // Parse details untuk mendapatkan informasi action
    if (details) {
      return details;
    }
    return action;
  };

  return (
    <Dropdown
      isOpen={isOpen}
      setIsOpen={(open) => {
        setIsOpen(open);

        if (setIsDotVisible) setIsDotVisible(false);
      }}
    >
      <DropdownTrigger
        className="grid size-12 place-items-center rounded-full border bg-gray-2 text-dark outline-none hover:text-primary focus-visible:border-primary focus-visible:text-primary dark:border-dark-4 dark:bg-dark-3 dark:text-white dark:focus-visible:border-primary"
        aria-label="View Notifications"
      >
        <span className="relative">
          <BellIcon />

          {isDotVisible && (
            <span
              className={cn(
                "absolute right-0 top-0 z-1 size-2 rounded-full bg-red-light ring-2 ring-gray-2 dark:ring-dark-3",
              )}
            >
              <span className="absolute inset-0 -z-1 animate-ping rounded-full bg-red-light opacity-75" />
            </span>
          )}
        </span>
      </DropdownTrigger>

      <DropdownContent
        align={isMobile ? "end" : "center"}
        className="border border-stroke bg-white px-3.5 py-3 shadow-md dark:border-dark-3 dark:bg-gray-dark min-[350px]:min-w-[20rem]"
      >
        <div className="mb-1 flex items-center justify-between px-2 py-1.5">
          <span className="text-lg font-medium text-dark dark:text-white">
            Activity Logs
          </span>
          {logs.length > 0 && (
            <span className="rounded-md bg-primary px-[9px] py-0.5 text-xs font-medium text-white">
              {logs.length} new
            </span>
          )}
        </div>

        <ul className="mb-3 max-h-[23rem] space-y-1.5 overflow-y-auto">
          {loading ? (
            <li className="px-2 py-4 text-center text-sm text-gray-6">
              Loading...
            </li>
          ) : logs.length === 0 ? (
            <li className="px-2 py-4 text-center text-sm text-gray-6">
              No activity logs
            </li>
          ) : (
            logs.map((log) => (
              <li key={log.id.toString()} role="menuitem">
                <div className="flex items-start gap-3 rounded-lg px-2 py-2 outline-none hover:bg-gray-2 focus-visible:bg-gray-2 dark:hover:bg-dark-3 dark:focus-visible:bg-dark-3">
                  <div className="mt-1 flex size-10 shrink-0 items-center justify-center rounded-full bg-primary/10 text-primary">
                    <span className="text-xs font-bold">
                      {log.user?.username?.[0]?.toUpperCase() || "A"}
                    </span>
                  </div>

                  <div className="min-w-0 flex-1">
                    <strong className="block text-sm font-medium text-dark dark:text-white">
                      {getActionText(log.action, log.details)}
                    </strong>

                    <span className="block text-xs text-dark-5 dark:text-dark-6">
                      by {log.user?.username || "Unknown"}
                    </span>

                    <span className="block text-xs text-dark-5 dark:text-dark-6">
                      {dayjs(log.createdAt).format("MMM DD, YYYY HH:mm")}
                    </span>
                  </div>
                </div>
              </li>
            ))
          )}
        </ul>
      </DropdownContent>
    </Dropdown>
  );
}