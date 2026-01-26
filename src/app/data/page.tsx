/**
 * Data Page
 * 
 * Halaman untuk menampilkan data pengguna dengan fitur:
 * - Pagination (default: 10 per page)
 * - Filter: Date range (default: hari ini)
 * - Bulk import dari Excel dengan template download
 * - Tampilan: Whatsapp, Name, NIK, Client, Created At
 * 
 * Data memiliki relasi dengan Client (bisa null untuk data tanpa client).
 */

import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { PaginationControls } from "@/components/pagination-controls";
import { DataImportForm } from "@/components/data-import-form";
import { DownloadButtonWrapper } from "@/components/DownloadButtonWrapper";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import {
  MIN_LIMIT,
  MAX_LIMIT,
  parseDateRangeParams,
  parsePaginationParams,
} from "@/lib/pagination";
import { prisma } from "@/lib/prisma";
import dayjs from "dayjs";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Data",
};

export const dynamic = "force-dynamic";
export const revalidate = 0;

type PageProps = {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
};

export default async function DataPage({ searchParams }: PageProps) {
  const resolved = await searchParams;
  const { page: rawPage, limit } = parsePaginationParams(resolved);
  const { startDate, endDate, startParam, endParam } =
    parseDateRangeParams(resolved);

  // Default to today if no date params
  const today = dayjs().format("YYYY-MM-DD");
  const defaultStartDate = startParam || today;
  const defaultEndDate = endParam || today;

  // Use today as default if no dates provided
  const filterStartDate = startDate || dayjs().startOf("day").toDate();
  const filterEndDate = endDate || dayjs().add(1, "day").startOf("day").toDate();

  const where = {
    createdAt: {
      gte: filterStartDate,
      lte: filterEndDate,
    },
  };

  const totalCount = await prisma.data.count({ where });
  const totalPages = Math.max(1, Math.ceil(totalCount / limit));
  const page = Math.min(rawPage, totalPages);
  const skip = (page - 1) * limit;

  const rows = await prisma.data.findMany({
    orderBy: { createdAt: "desc" },
    skip,
    take: limit,
    where,
    select: {
      id: true,
      whatsapp: true,
      name: true,
      nik: true,
      createdAt: true,
      client: {
        select: {
          name: true,
        },
      },
    },
  });

  return (
    <div className="space-y-6">
      <Breadcrumb pageName="Data" />

      <DataImportForm />

      <div className="rounded-[10px] border border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card sm:p-7.5">
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-semibold text-dark dark:text-white">
              Data
            </h3>
            <p className="text-sm text-neutral-500 dark:text-neutral-300">
              Menampilkan data pengguna dengan relasi client
            </p>
          </div>

          <div className="text-sm text-neutral-500 dark:text-neutral-300">
            Limit {limit} · Page {page} of {totalPages}
          </div>
        </div>

        <form className="mb-4 flex flex-wrap items-end gap-3 text-sm" method="get">
          <div className="flex flex-col gap-1">
            <label className="text-neutral-600 dark:text-neutral-300" htmlFor="start">
              Start date
            </label>
            <input
              id="start"
              name="start"
              type="date"
              defaultValue={defaultStartDate}
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
              defaultValue={defaultEndDate}
              className="rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
            />
          </div>

          <div className="flex flex-col gap-1">
            <label className="text-neutral-600 dark:text-neutral-300" htmlFor="limit">
              Show
            </label>
            <select
              id="limit"
              name="limit"
              defaultValue={String(limit)}
              className="h-10 w-36 rounded-md border border-stroke px-3 text-sm dark:border-dark-3 dark:bg-dark-2"
            >
              {[10, 25, 50, 100, 250, 500, 1000, 5000, 10000].map((v) => (
                <option key={v} value={String(v)}>
                  {v}
                </option>
              ))}
            </select>
          </div>

          <input type="hidden" name="page" value="1" />

          <button
            type="submit"
            className="h-10 rounded-md bg-primary px-4 font-medium text-white transition hover:bg-opacity-90"
          >
            Terapkan
          </button>

          <a
            href="/data"
            className="h-10 rounded-md border border-stroke px-4 font-medium leading-10 text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:text-white"
          >
            Reset
          </a>

          <DownloadButtonWrapper type="data" />
        </form>

        <Table>
          <TableHeader>
            <TableRow className="border-none bg-[#F7F9FC] dark:bg-dark-2 [&>th]:py-4 [&>th]:text-base [&>th]:text-dark [&>th]:dark:text-white">
              <TableHead className="min-w-[160px]">Whatsapp</TableHead>
              <TableHead className="min-w-[160px]">Name</TableHead>
              <TableHead className="min-w-[160px]">NIK</TableHead>
              <TableHead className="min-w-[160px]">Client</TableHead>
              <TableHead className="min-w-[160px]">Created At</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {rows.length === 0 ? (
              <TableRow>
                <TableCell
                  className="text-center text-neutral-500 dark:text-neutral-300"
                  colSpan={5}
                >
                  Belum ada data.
                </TableCell>
              </TableRow>
            ) : (
              rows.map((item) => (
                <TableRow
                  key={item.id.toString()}
                  className="border-[#eee] dark:border-dark-3"
                >
                  <TableCell className="font-medium text-dark dark:text-white">
                    {item.whatsapp || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {item.name || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {item.nik || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {item.client?.name || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {item.createdAt
                      ? dayjs(item.createdAt).format("YYYY-MM-DD HH:mm")
                      : "-"}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <PaginationControls
          basePath="/data"
          limit={limit}
          page={page}
          total={totalCount}
          params={{ start: startParam, end: endParam }}
        />
      </div>
    </div>
  );
}

