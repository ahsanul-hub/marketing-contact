"use client";

import { useRouter, useSearchParams } from "next/navigation";
import dayjs from "dayjs";

type QuickFilterType = "today" | "yesterday" | "7d" | "30d" | "month";

export function QuickFilters() {
  const router = useRouter();
  const searchParams = useSearchParams();

  const getDateRangeForFilter = (filter: QuickFilterType) => {
    const today = dayjs();
    let start: dayjs.Dayjs;
    let end: dayjs.Dayjs = today;

    switch (filter) {
      case "today":
        start = today;
        break;
      case "yesterday":
        start = today.subtract(1, "day");
        end = today.subtract(1, "day");
        break;
      case "7d":
        start = today.subtract(6, "day");
        break;
      case "30d":
        start = today.subtract(29, "day");
        break;
      case "month":
        start = today.startOf("month");
        end = today.endOf("month");
        break;
      default:
        start = today;
    }

    return {
      startDate: start.format("YYYY-MM-DD"),
      endDate: end.format("YYYY-MM-DD"),
    };
  };

  const handleQuickFilter = (filter: QuickFilterType) => {
    const { startDate, endDate } = getDateRangeForFilter(filter);
    const params = new URLSearchParams(searchParams);
    params.set("start", startDate);
    params.set("end", endDate);
    router.push(`/?${params.toString()}`);
  };

  const getActiveFilter = (): QuickFilterType | null => {
    const start = searchParams.get("start");
    const end = searchParams.get("end");

    if (!start || !end) return null;

    const today = dayjs().format("YYYY-MM-DD");
    const yesterday = dayjs().subtract(1, "day").format("YYYY-MM-DD");
    const weekAgo = dayjs().subtract(6, "day").format("YYYY-MM-DD");
    const monthAgo = dayjs().subtract(29, "day").format("YYYY-MM-DD");
    const monthStart = dayjs().startOf("month").format("YYYY-MM-DD");
    const monthEnd = dayjs().endOf("month").format("YYYY-MM-DD");

    if (start === today && end === today) return "today";
    if (start === yesterday && end === yesterday) return "yesterday";
    if (start === weekAgo && end === today) return "7d";
    if (start === monthAgo && end === today) return "30d";
    if (start === monthStart && end === monthEnd) return "month";

    return null;
  };

  const activeFilter = getActiveFilter();

  const buttonClass = (filter: QuickFilterType) => {
    const isActive = activeFilter === filter;
    return `px-4 py-2 rounded-md text-sm font-medium transition ${
      isActive
        ? "bg-primary text-white"
        : "bg-white dark:bg-gray-dark border border-stroke dark:border-dark-3 text-dark dark:text-white hover:bg-gray-1 dark:hover:bg-gray-dark"
    }`;
  };

  return (
    <div className="flex flex-wrap gap-2">
      <button
        onClick={() => handleQuickFilter("today")}
        className={buttonClass("today")}
      >
        Today
      </button>
      <button
        onClick={() => handleQuickFilter("yesterday")}
        className={buttonClass("yesterday")}
      >
        Yesterday
      </button>
      <button
        onClick={() => handleQuickFilter("7d")}
        className={buttonClass("7d")}
      >
        7D
      </button>
      <button
        onClick={() => handleQuickFilter("30d")}
        className={buttonClass("30d")}
      >
        30D
      </button>
      <button
        onClick={() => handleQuickFilter("month")}
        className={buttonClass("month")}
      >
        This Month
      </button>
    </div>
  );
}
