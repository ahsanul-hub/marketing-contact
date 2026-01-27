/**
 * Registration Page
 * 
 * Halaman untuk menampilkan data registrasi dengan fitur:
 * - Pagination (default: 10 per page, bisa diubah)
 * - Filter: Date range (default: hari ini), Client (All/Organic/Specific)
 * - Bulk import dari Excel dengan template download
 * - Tampilan tabel dengan phone number dan created at
 * 
 * Query Logic:
 * - Organic = phone_number yang tidak ada di tabel data (whatsapp)
 * - Client filter = phone_number yang ada di data dengan client tertentu
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
  const isOrganic = clientIdParam === "organic";
  const clientId =
    clientIdParam && clientIdParam !== "organic" ? BigInt(clientIdParam) : undefined;

  // Dropdown list for client filter
  const clients = await prisma.client.findMany({
    orderBy: [{ name: "asc" }, { id: "asc" }],
    select: { id: true, name: true },
  });

  // NOTE:
  // User requested: "organic = phone_number tidak ada di tabel data kolom phone_number".
  // In current schema, Data does not have phone_number; we use Data.whatsapp as the phone identifier.
  // Use today as default if no dates provided
  const filterStartDate = startDate
    ? dayjs(startDate).startOf("day").toDate()
    : dayjs().startOf("day").toDate();
  const filterEndDate = endDate
    ? dayjs(endDate).endOf("day").toDate()
    : dayjs().endOf("day").toDate();
  
  const dateFilterSql = Prisma.sql` AND r.created_at >= ${filterStartDate} AND r.created_at <= ${filterEndDate}`;

  const typeFilterSql =
    isOrganic
      ? Prisma.sql` AND NOT EXISTS (
          SELECT 1 FROM data d
          WHERE d.whatsapp = r.phone_number
        )`
      : clientId
        ? Prisma.sql` AND EXISTS (
            SELECT 1
            FROM data d
            WHERE d.whatsapp = r.phone_number
              AND d.id_client = ${clientId}
          )`
        : Prisma.empty;

  const totalCountRow = await prisma.$queryRaw<{ count: bigint }[]>`
    SELECT COUNT(*)::bigint as count
    FROM registration r
    WHERE 1=1
      ${dateFilterSql}
      ${typeFilterSql}
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
      ${typeFilterSql}
    ORDER BY r.created_at DESC NULLS LAST, r.id DESC
    LIMIT ${limit} OFFSET ${skip}
  `;

  return (
    <div className="space-y-6">
      <Breadcrumb pageName="Registration" />

      <RegistrationImportForm />

      <DownloadButtonWrapper type="registration" />

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
            <label className="text-neutral-600 dark:text-neutral-300" htmlFor="client_id">
              Client
            </label>
            <select
              id="client_id"
              name="client_id"
              defaultValue={clientIdParam || ""}
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
          }}
        />
      </div>
    </div>
  );
}
