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
import { Prisma } from "@prisma/client";

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

  const searchParam = resolved?.search ? String(resolved.search) : undefined;

  const filterStartDate = startDate
    ? dayjs(startDate).startOf("day").toDate()
    : dayjs().startOf("day").toDate();
  const filterEndDate = endDate
    ? dayjs(endDate).endOf("day").toDate()
    : dayjs().endOf("day").toDate();

  const dateFilterSql = Prisma.sql` AND d.created_at >= ${filterStartDate} AND d.created_at <= ${filterEndDate}`;

  const searchFilterSql = searchParam
    ? Prisma.sql` AND (
        d.whatsapp ILIKE ${`%${searchParam}%`} OR
        d.name ILIKE ${`%${searchParam}%`} OR
        d.nik ILIKE ${`%${searchParam}%`} OR
        c.name ILIKE ${`%${searchParam}%`}
      )`
    : Prisma.empty;

  const totalCountRow = await prisma.$queryRaw<{ count: bigint }[]>`
    SELECT COUNT(*)::bigint as count
    FROM data d
    LEFT JOIN client c ON d.id_client = c.id
    WHERE 1=1
      ${dateFilterSql}
      ${searchFilterSql}
  `;

  const totalCount = Number(totalCountRow?.[0]?.count ?? Number(0));
  const totalPages = Math.max(1, Math.ceil(totalCount / limit));
  const page = Math.min(rawPage, totalPages);
  const skip = (page - 1) * limit;

  const rows = await prisma.$queryRaw<
    { id: bigint; whatsapp: string | null; name: string | null; nik: string | null; created_at: Date | null; client_name: string | null }[]
  >`
    SELECT d.id, d.whatsapp, d.name, d.nik, d.created_at, c.name as client_name
    FROM data d
    LEFT JOIN client c ON d.id_client = c.id
    WHERE 1=1
      ${dateFilterSql}
      ${searchFilterSql}
    ORDER BY d.created_at DESC NULLS LAST, d.id DESC
    LIMIT ${limit} OFFSET ${skip}
  `;

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

          <div className="flex flex-col gap-1">
            <label className="text-neutral-600 dark:text-neutral-300" htmlFor="search">
              Search
            </label>
            <input
              id="search"
              name="search"
              type="text"
              placeholder="Whatsapp, Name, NIK or Client"
              defaultValue={searchParam || ""}
              className="h-10 w-56 rounded-md border border-stroke px-3 text-sm dark:border-dark-3 dark:bg-dark-2"
            />
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
                    {item.client_name || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {item.created_at
                      ? dayjs(item.created_at).format("YYYY-MM-DD HH:mm")
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
          params={{ start: startParam, end: endParam, search: searchParam }}
        />
      </div>
    </div>
  );
}

