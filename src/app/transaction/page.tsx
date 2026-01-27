/**
 * Transaction Page
 * 
 * Halaman untuk menampilkan data transaksi dengan fitur:
 * - Pagination (default: 10 per page)
 * - Filter: Date range (default: hari ini)
 * - Bulk import dari Excel dengan template download
 * - Tampilan: Transaction Date, Phone Number, Total Deposit, Total Profit
 * 
 * Data diurutkan berdasarkan transaction_date DESC (terbaru dulu).
 */

import Breadcrumb from "@/components/Breadcrumbs/Breadcrumb";
import { PaginationControls } from "@/components/pagination-controls";
import { TransactionImportForm } from "@/components/transaction-import-form";
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
  title: "Transaction",
};

export const dynamic = "force-dynamic";
export const revalidate = 0;

type PageProps = {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
};

const numberFormatter = new Intl.NumberFormat("en-US");

export default async function TransactionPage({ searchParams }: PageProps) {
  const resolved = await searchParams;
  const { page: rawPage, limit } = parsePaginationParams(resolved);
  const { startDate, endDate, startParam, endParam } =
    parseDateRangeParams(resolved);

  // Default to today if no date params
  const today = dayjs().format("YYYY-MM-DD");
  const defaultStartDate = startParam || today;
  const defaultEndDate = endParam || today;

  const searchParam = resolved?.search ? String(resolved.search) : undefined;

  // Use today as default if no dates provided
  const filterStartDate = startDate
    ? dayjs(startDate).startOf("day").toDate()
    : dayjs().startOf("day").toDate();
  const filterEndDate = endDate
    ? dayjs(endDate).endOf("day").toDate()
    : dayjs().add(1, "day").startOf("day").toDate();

  const where = {
    transactionDate: {
      gte: filterStartDate,
      lte: filterEndDate,
    },
    ...(searchParam && {
      OR: [
        { phoneNumber: { contains: searchParam, mode: "insensitive" } },
        { client: { name: { contains: searchParam, mode: "insensitive" } } },
      ],
    }),
  };

  const totalCount = await prisma.transaction.count({ where });
  const totalPages = Math.max(1, Math.ceil(totalCount / limit));
  const page = Math.min(rawPage, totalPages);
  const skip = (page - 1) * limit;

  const transactions = await prisma.transaction.findMany({
    orderBy: { transactionDate: "desc" },
    skip,
    take: limit,
    where,
    select: {
      id: true,
      phoneNumber: true,
      transactionDate: true,
      totalDeposit: true,
      totalProfit: true,
      client: {
        select: {
          name: true,
        },
      },
    },
  });

  return (
    <div className="space-y-6">
      <Breadcrumb pageName="Transaction" />

      <TransactionImportForm />

      <DownloadButtonWrapper type="transaction" />

      <div className="rounded-[10px] border border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card sm:p-7.5">
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-semibold text-dark dark:text-white">
              Transaction
            </h3>
            <p className="text-sm text-neutral-500 dark:text-neutral-300">
              Latest transactions ordered by newest first
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
            href="/transaction"
            className="h-10 rounded-md border border-stroke px-4 font-medium leading-10 text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:text-white"
          >
            Reset
          </a>

          <DownloadButtonWrapper type="transaction" />
        </form>

        <Table>
          <TableHeader>
            <TableRow className="border-none bg-[#F7F9FC] dark:bg-dark-2 [&>th]:py-4 [&>th]:text-base [&>th]:text-dark [&>th]:dark:text-white">
              <TableHead className="min-w-[180px]">Transaction Date</TableHead>
              <TableHead className="min-w-[180px]">Phone Number</TableHead>
              <TableHead className="min-w-[140px]">Total Deposit</TableHead>
              <TableHead className="min-w-[140px]">Total Profit</TableHead>
              <TableHead className="min-w-[140px]">Client</TableHead>
            </TableRow>
          </TableHeader>

        <TableBody>
            {transactions.length === 0 ? (
              <TableRow>
                <TableCell
                  className="text-center text-neutral-500 dark:text-neutral-300"
                  colSpan={5}
                >
                  Belum ada data transaksi.
                </TableCell>
              </TableRow>
            ) : (
              transactions.map((item) => (
                <TableRow
                  key={item.id.toString()}
                  className="border-[#eee] dark:border-dark-3"
                >
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {dayjs(item.transactionDate).format("YYYY-MM-DD HH:mm")}
                  </TableCell>
                  <TableCell className="font-medium text-dark dark:text-white">
                    {item.phoneNumber || "-"}
                  </TableCell>
                  <TableCell className="text-dark dark:text-white">
                    {numberFormatter.format(
                      item.totalDeposit ? Number(item.totalDeposit) : 0,
                    )}
                  </TableCell>
                  <TableCell className="text-dark dark:text-white">
                    {numberFormatter.format(
                      item.totalProfit ? Number(item.totalProfit) : 0,
                    )}
                  </TableCell>
                  <TableCell className="text-dark dark:text-white">
                    {item.client?.name || "-"}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <PaginationControls
          basePath="/transaction"
          limit={limit}
          page={page}
          total={totalCount}
          params={{ start: startParam, end: endParam, search: searchParam }}
        />
      </div>
    </div>
  );
}
