"use client";

import { downloadDataTemplate, downloadRegistrationTemplate, downloadTransactionTemplate } from "@/lib/excel-template";

type TemplateType = "registration" | "transaction" | "data";

interface TemplateDownloadButtonProps {
  type: TemplateType;
}

export function TemplateDownloadButton({ type }: TemplateDownloadButtonProps) {
  const handleDownload = () => {
    switch (type) {
      case "registration":
        downloadRegistrationTemplate();
        break;
      case "transaction":
        downloadTransactionTemplate();
        break;
      case "data":
        downloadDataTemplate();
        break;
    }
  };

  return (
    <button
      type="button"
      onClick={handleDownload}
      className="flex items-center gap-2 rounded-md border border-stroke bg-white px-4 py-2 text-sm font-medium text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:bg-dark-2 dark:text-white dark:hover:border-primary dark:hover:text-primary"
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
      Download Template
    </button>
  );
}
