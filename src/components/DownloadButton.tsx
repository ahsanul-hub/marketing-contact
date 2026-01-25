"use client";

import { useSearchParams } from "next/navigation";
import { useState, useRef, useEffect } from "react";

interface DownloadButtonProps {
  type: "registration" | "transaction" | "data" | "home";
  className?: string;
}

export function DownloadButton({ type, className }: DownloadButtonProps) {
  const searchParams = useSearchParams();
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }

    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  const handleDownload = (format: "csv" | "xlsx") => {
    const params = new URLSearchParams(searchParams);
    // Remove pagination params for full export
    params.delete("page");
    params.delete("limit");
    
    // Add format parameter (only for data type, others default to xlsx)
    if (type === "data" && format === "csv") {
      params.set("format", "csv");
    } else if (type === "data" && format === "xlsx") {
      params.set("format", "xlsx");
    }

    const url = `/api/export/${type}?${params.toString()}`;
    window.open(url, "_blank");
    setIsOpen(false);
  };

  // For non-data types, just show single Excel button
  if (type !== "data") {
    return (
      <button
        type="button"
        onClick={() => handleDownload("xlsx")}
        className={`flex items-center gap-2 rounded-md border border-stroke bg-white px-4 py-2 text-sm font-medium text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:bg-dark-2 dark:text-white dark:hover:border-primary dark:hover:text-primary ${className || ""}`}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
          stroke="currentColor"
          className="h-4 w-4"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3"
          />
        </svg>
        Download Excel
      </button>
    );
  }

  // For data type, show dropdown with CSV and Excel options
  return (
    <div className="relative" ref={dropdownRef}>
      <button
        type="button"
        onClick={() => setIsOpen(!isOpen)}
        className={`flex items-center gap-2 rounded-md border border-stroke bg-white px-4 py-2 text-sm font-medium text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:bg-dark-2 dark:text-white dark:hover:border-primary dark:hover:text-primary ${className || ""}`}
      >
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
          stroke="currentColor"
          className="h-4 w-4"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3"
          />
        </svg>
        Download
        <svg
          xmlns="http://www.w3.org/2000/svg"
          fill="none"
          viewBox="0 0 24 24"
          strokeWidth={1.5}
          stroke="currentColor"
          className="h-3 w-3"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            d="m19.5 8.25-7.5 7.5-7.5-7.5"
          />
        </svg>
      </button>

      {isOpen && (
        <div className="absolute right-0 z-10 mt-2 w-48 rounded-md border border-stroke bg-white shadow-lg dark:border-dark-3 dark:bg-gray-dark">
          <button
            type="button"
            onClick={() => handleDownload("csv")}
            className="flex w-full items-center gap-2 px-4 py-2 text-sm text-dark transition hover:bg-gray-100 dark:text-white dark:hover:bg-dark-2"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="h-4 w-4"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3"
              />
            </svg>
            Download CSV
          </button>
          <button
            type="button"
            onClick={() => handleDownload("xlsx")}
            className="flex w-full items-center gap-2 px-4 py-2 text-sm text-dark transition hover:bg-gray-100 dark:text-white dark:hover:bg-dark-2"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              fill="none"
              viewBox="0 0 24 24"
              strokeWidth={1.5}
              stroke="currentColor"
              className="h-4 w-4"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                d="M3 16.5v2.25A2.25 2.25 0 005.25 21h13.5A2.25 2.25 0 0021 18.75V16.5M16.5 12L12 16.5m0 0L7.5 12m4.5 4.5V3"
              />
            </svg>
            Download Excel
          </button>
        </div>
      )}
    </div>
  );
}