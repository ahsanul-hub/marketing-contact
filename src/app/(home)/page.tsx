/**
 * Dashboard Home Page
 * 
 * Halaman utama dashboard yang menampilkan:
 * - Overview cards: Total Deposit, Total Profit, Registrations, Contacts
 * - Charts: Payments Overview, Weeks Profit, Clients Distribution
 * - Tables: Top Profit, Top 10 Clients
 * 
 * Filter yang tersedia:
 * - Date range (start date, end date) - default: hari ini
 * - Client filter (All, Organic, atau specific client)
 * 
 * Semua data di-fetch dari database menggunakan Prisma.
 */

import { PaymentsOverview } from "@/components/Charts/payments-overview";
import { UsedDevices } from "@/components/Charts/used-devices";
import { WeeksProfit } from "@/components/Charts/weeks-profit";
import { TopChannels } from "@/components/Tables/top-channels";
import { TopChannelsSkeleton } from "@/components/Tables/top-channels/skeleton";
import { TopClients } from "@/components/Charts/top-clients";
import { createTimeFrameExtractor } from "@/utils/timeframe-extractor";
import { parseDateRangeParams } from "@/lib/pagination";
import { prisma } from "@/lib/prisma";
import { Suspense } from "react";
import { OverviewCardsGroup } from "./_components/overview-cards";
import { OverviewCardsSkeleton } from "./_components/overview-cards/skeleton";
import { DownloadButtonWrapper } from "@/components/DownloadButtonWrapper";

type PropsType = {
  searchParams: Promise<{
    selected_time_frame?: string;
    start?: string;
    end?: string;
    client_id?: string;
  }>;
};

export default async function Home({ searchParams }: PropsType) {
  const { selected_time_frame, start, end, client_id } = await searchParams;
  const extractTimeFrame = createTimeFrameExtractor(selected_time_frame);

  const { startDate, endDate, startParam, endParam } = parseDateRangeParams({
    start,
    end,
  });

  const clients = await prisma.client.findMany({
    orderBy: [{ name: "asc" }, { id: "asc" }],
    select: {
      id: true,
      name: true,
    },
  });

  const isOrganic = client_id === "organic";
  const clientFilterId =
    client_id && client_id !== "organic" ? BigInt(client_id) : undefined;

  const analyticsFilter = {
    startDate,
    endDate,
    clientId: clientFilterId,
    isOrganic,
  } as const;

  return (
    <>
      <Suspense fallback={<OverviewCardsSkeleton />}>
        <OverviewCardsGroup {...analyticsFilter} />
      </Suspense>

      <form className="mt-4 mb-4 flex flex-wrap items-end gap-3 rounded-[10px] border border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card text-sm">
        <div className="flex flex-col gap-1">
          <label className="text-neutral-600 dark:text-neutral-300" htmlFor="start">
            Start date
          </label>
          <input
            id="start"
            name="start"
            type="date"
            defaultValue={startParam}
            className="rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
          />
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-neutral-600 dark:text-neutral-300" htmlFor="end">
            End date
          </label>
          <input
            id="end"
            name="end"
            type="date"
            defaultValue={endParam}
            className="rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
          />
        </div>

        <div className="flex flex-col gap-1">
          <label className="text-neutral-600 dark:text-neutral-300" htmlFor="client_id">
            Client
          </label>
          <select
            id="client_id"
            name="client_id"
            defaultValue={client_id || ""}
            className="h-10 w-56 rounded-md border border-stroke px-3 text-sm dark:border-dark-3 dark:bg-dark-2"
          >
            <option value="">All</option>
            <option value="organic">Organic</option>
            {clients.map((c) => (
              <option key={c.id.toString()} value={c.id.toString()}>
                {c.name || `Client #${c.id.toString()}`}
              </option>
            ))}
          </select>
        </div>

        <button
          type="submit"
          className="h-10 rounded-md bg-primary px-4 font-medium text-white transition hover:bg-opacity-90"
        >
          Terapkan
        </button>

        <a
          href="/"
          className="h-10 rounded-md border border-stroke px-4 font-medium leading-10 text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:text-white"
        >
          Reset
        </a>

        <DownloadButtonWrapper type="home" />
      </form>

      <div className="mt-4 grid grid-cols-12 gap-4 md:mt-6 md:gap-6 2xl:mt-9 2xl:gap-7.5">
        <PaymentsOverview
          className="col-span-12 xl:col-span-7"
          key={extractTimeFrame("payments_overview")}
          timeFrame={extractTimeFrame("payments_overview")?.split(":")[1]}
          filter={analyticsFilter}
        />

        <WeeksProfit
          key={extractTimeFrame("weeks_profit")}
          timeFrame={extractTimeFrame("weeks_profit")?.split(":")[1]}
          filter={analyticsFilter}
          className="col-span-12 xl:col-span-5"
        />

        <UsedDevices
          className="col-span-12 xl:col-span-5"
          key={extractTimeFrame("used_devices")}
          timeFrame={extractTimeFrame("used_devices")?.split(":")[1]}
        />

        <div className="col-span-12 grid xl:col-span-8">
          <Suspense fallback={<TopChannelsSkeleton />}>
            <TopChannels />
          </Suspense>
        </div>

        <Suspense fallback={null}>
          <TopClients className="col-span-12 xl:col-span-4" />
        </Suspense>
      </div>
    </>
  );
}
