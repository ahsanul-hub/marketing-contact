import "@/css/satoshi.css";
import "@/css/style.css";

import "flatpickr/dist/flatpickr.min.css";
import "jsvectormap/dist/jsvectormap.css";

import NextTopLoader from "nextjs-toploader";
import type { PropsWithChildren } from "react";
import { Providers } from "../providers";

export default function AuthLayout({ children }: PropsWithChildren) {
  return (
    <div className="flex min-h-screen bg-gray-2 dark:bg-[#020d1a]">
      <div className="w-full">
        <div className="isolate mx-auto w-full max-w-screen-2xl overflow-hidden p-4 md:p-6 2xl:p-10">
          {children}
        </div>
      </div>
    </div>
  );
}