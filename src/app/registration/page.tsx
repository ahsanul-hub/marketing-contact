/**
 * Registration Page
 * 
 * Halaman untuk menampilkan data registrasi dengan fitur:
 * - Pagination (default: 10 per page, bisa diubah)
 * - Filter: Date range (default: hari ini), Organic/Non-Organic, Client (specific client)
 * - Bulk import dari Excel dengan template download
 * - Tampilan tabel dengan phone number dan created at
 * 
 * Query Logic:
 * - Organic = phone_number yang tidak ada di tabel data (whatsapp)
 * - Non-Organic = phone_number yang ada di tabel data
 * - Client filter = Non-Organic registrations linked to specific client
 */

import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { PaginationControls } from "@/components/pagination-controls";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { RegistrationImportForm } from "@/components/registration-import-form";
import { DownloadButtonWrapper } from "@/components/DownloadButtonWrapper";
import { DownloadButton } from "@/components/DownloadButton";
import { RegistrationActions } from "./_components/registration-actions";
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
  title: "Registration",
};

export const dynamic = "force-dynamic";
export const revalidate = 0;

type PageProps = {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
};

function normalizeParam(param: string | string[] | undefined) {
  if (!param) return undefined;
  return Array.isArray(param) ? param[0] : param;
}

export default async function RegistrationPage({ searchParams }: PageProps) {
  const resolved = await searchParams;
  const { page: rawPage, limit } = parsePaginationParams(resolved);
  const { startDate, endDate, startParam, endParam } =
    parseDateRangeParams(resolved);

  // Default to today if no date params
  const today = dayjs().format("YYYY-MM-DD");
  const defaultStartDate = startParam || today;
  const defaultEndDate = endParam || today;

  const clientIdParam = normalizeParam(resolved?.client_id);
  const organicParam = normalizeParam(resolved?.organic);
  const searchParam = normalizeParam(resolved?.search);
  
  // Organic filter: "all" | "organic" | "non-organic" (default: "all")
  const organicType = organicParam || "all";
  const clientId = clientIdParam ? Number(clientIdParam) : undefined;

  // Dropdown list for client filter
  const clients = await prisma.client.findMany({
    orderBy: [{ name: "asc" }, { id: "asc" }],
    select: { id: true, name: true },
  });

  // Use today as default if no dates provided
  const filterStartDate = startDate
    ? dayjs(startDate).startOf("day").toDate()
    : dayjs().startOf("day").toDate();
  const filterEndDate = endDate
    ? dayjs(endDate).endOf("day").toDate()
    : dayjs().endOf("day").toDate();
  
  const dateFilterSql = Prisma.sql` AND r.created_at >= ${filterStartDate} AND r.created_at <= ${filterEndDate}`;

  const searchFilterSql = searchParam
    ? Prisma.sql` AND (
        r.phone_number ILIKE ${`%${searchParam}%`} OR
        c.name ILIKE ${`%${searchParam}%`}
      )`
    : Prisma.empty;

  // Organic/Non-Organic filter
  const organicFilterSql =
    organicType === "organic"
      ? Prisma.sql` AND NOT EXISTS (
          SELECT 1 FROM data d
          WHERE d.whatsapp = r.phone_number
        )`
      : organicType === "non-organic"
        ? Prisma.sql` AND EXISTS (
            SELECT 1 FROM data d
            WHERE d.whatsapp = r.phone_number
          )`
        : Prisma.empty;


  const totalCountRow = await prisma.$queryRaw<{ count: number }[]>`
    SELECT COUNT(*)::int as count
    FROM registration r
    LEFT JOIN client c ON r.id_client = c.id
    WHERE 1=1
      ${dateFilterSql}
      ${searchFilterSql}
      ${organicFilterSql}
  `;

  const totalCount = Number(totalCountRow?.[0]?.count ?? Number(0));
  const totalPages = Math.max(1, Math.ceil(totalCount / limit));
  const page = Math.min(rawPage, totalPages);
  const skip = (page - 1) * limit;

  const registrations = await prisma.$queryRaw<
    { id: bigint; phone_number: string | null; created_at: Date | null; client_name: string | null }[]
  >`
    SELECT r.id, r.phone_number, r.created_at, c.name as client_name
    FROM registration r
    LEFT JOIN client c ON r.id_client = c.id
    WHERE 1=1
      ${dateFilterSql}
      ${searchFilterSql}
      ${organicFilterSql}
    ORDER BY r.created_at DESC NULLS LAST, r.id DESC
    LIMIT ${limit} OFFSET ${skip}
  `;

  return (
    <div className="space-y-6">
      <Breadcrumb pageName="Registration" />

      <RegistrationImportForm />

      <div className="rounded-[10px] border border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card sm:p-7.5">
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-semibold text-dark dark:text-white">
              Registration
            </h3>
            <p className="text-sm text-neutral-500 dark:text-neutral-300">
              Phone numbers ordered by newest first
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
            <label className="text-neutral-600 dark:text-neutral-300" htmlFor="organic">
              Type
            </label>
            <select
              id="organic"
              name="organic"
              defaultValue={organicParam || "all"}
              className="h-10 w-48 rounded-md border border-stroke px-3 text-sm dark:border-dark-3 dark:bg-dark-2"
            >
              <option value="all">All</option>
              <option value="organic">Organic</option>
              <option value="non-organic">Non-Organic</option>
            </select>
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
              placeholder="Phone or Client"
              defaultValue={searchParam || ""}
              className="h-10 w-48 rounded-md border border-stroke px-3 text-sm dark:border-dark-3 dark:bg-dark-2"
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
            href="/registration"
            className="h-10 rounded-md border border-stroke px-4 font-medium leading-10 text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:text-white"
          >
            Reset
          </a>

          <DownloadButtonWrapper type="registration" />
        </form>

        <Table>
          <TableHeader>
            <TableRow className="border-none bg-[#F7F9FC] dark:bg-dark-2 [&>th]:py-4 [&>th]:text-base [&>th]:text-dark [&>th]:dark:text-white">
              <TableHead className="min-w-[220px]">Phone Number</TableHead>
              <TableHead className="min-w-[180px]">Created At</TableHead>
              <TableHead className="min-w-[140px]">Client</TableHead>
              <TableHead className="min-w-[120px]">Actions</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {registrations.length === 0 ? (
              <TableRow>
                <TableCell
                  className="text-center text-neutral-500 dark:text-neutral-300"
                  colSpan={3}
                >
                  Belum ada data registrasi.
                </TableCell>
              </TableRow>
            ) : (
              registrations.map((item) => (
                <TableRow
                  key={item.id.toString()}
                  className="border-[#eee] dark:border-dark-3"
                >
                  <TableCell className="font-medium text-dark dark:text-white">
                    {item.phone_number || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {item.created_at
                      ? dayjs(item.created_at).format("YYYY-MM-DD HH:mm")
                      : "-"}
                  </TableCell>
                  <TableCell className="text-dark dark:text-white">
                    {item.client_name || "-"}
                  </TableCell>
                  <TableCell>
                    <RegistrationActions
                      registrationId={item.id}
                      phoneNumber={item.phone_number || ""}
                      clientName={item.client_name || undefined}
                      createdAt={item.created_at ? item.created_at.toISOString() : undefined}
                    />
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <PaginationControls
          basePath="/registration"
          limit={limit}
          page={page}
          total={totalCount}
          params={{
            start: startParam,
            end: endParam,
            client_id: clientIdParam,
            search: searchParam,
          }}
        />
      </div>
    </div>
  );
}
