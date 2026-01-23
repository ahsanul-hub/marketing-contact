"use client";

import { DownloadButton } from "@/components/DownloadButton";

interface DownloadButtonWrapperProps {
  type: "registration" | "transaction" | "data" | "home";
  className?: string;
}

export function DownloadButtonWrapper({ type, className }: DownloadButtonWrapperProps) {
  return <DownloadButton type={type} className={className} />;
}