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
import {
  MIN_LIMIT,
  MAX_LIMIT,
  parseDateRangeParams,
  parsePaginationParams,
} from "@/lib/pagination";
import { prisma } from "@/lib/prisma";
import dayjs from "dayjs";
import type { Metadata } from "next";
import { CreateClientForm } from "./_components/create-client-form";

export const metadata: Metadata = {
  title: "Client",
};

export const dynamic = "force-dynamic";
export const revalidate = 0;

type PageProps = {
  searchParams?: Promise<{ [key: string]: string | string[] | undefined }>;
};

export default async function ClientPage({ searchParams }: PageProps) {
  const resolved = await searchParams;
  const { page: rawPage, limit } = parsePaginationParams(resolved);
  const { startDate, endDate, startParam, endParam } =
    parseDateRangeParams(resolved);

  const where =
    startDate || endDate
      ? {
          createdAt: {
            gte: startDate,
            lte: endDate,
          },
        }
      : undefined;

  const totalCount = await prisma.client.count({ where });
  const totalPages = Math.max(1, Math.ceil(totalCount / limit));
  const page = Math.min(rawPage, totalPages);
  const skip = (page - 1) * limit;

  const clients = await prisma.client.findMany({
    orderBy: [
      { createdAt: "desc" },
      { id: "desc" },
    ],
    skip,
    take: limit,
    where,
    select: {
      id: true,
      name: true,
      createdAt: true,
    },
  });

  return (
    <div className="space-y-6">
      <Breadcrumb pageName="Client" />

      <CreateClientForm />

      <div className="rounded-[10px] border border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card sm:p-7.5">
        <div className="mb-4 flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-semibold text-dark dark:text-white">
              Client
            </h3>
            <p className="text-sm text-neutral-500 dark:text-neutral-300">
              Daftar client, urut terbaru
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
            href="/client"
            className="h-10 rounded-md border border-stroke px-4 font-medium leading-10 text-dark transition hover:border-primary hover:text-primary dark:border-dark-3 dark:text-white"
          >
            Reset
          </a>
        </form>

        <Table>
          <TableHeader>
            <TableRow className="border-none bg-[#F7F9FC] dark:bg-dark-2 [&>th]:py-4 [&>th]:text-base [&>th]:text-dark [&>th]:dark:text-white">
              <TableHead className="min-w-[200px]">Client Name</TableHead>
              <TableHead className="min-w-[200px]">Created At</TableHead>
            </TableRow>
          </TableHeader>

          <TableBody>
            {clients.length === 0 ? (
              <TableRow>
                <TableCell
                  className="text-center text-neutral-500 dark:text-neutral-300"
                  colSpan={2}
                >
                  Belum ada client.
                </TableCell>
              </TableRow>
            ) : (
              clients.map((client) => (
                <TableRow
                  key={client.id.toString()}
                  className="border-[#eee] dark:border-dark-3"
                >
                  <TableCell className="font-medium text-dark dark:text-white">
                    {client.name || "-"}
                  </TableCell>
                  <TableCell className="text-neutral-600 dark:text-neutral-300">
                    {client.createdAt
                      ? dayjs(client.createdAt).format("DD-MM-YYYY HH:mm")
                      : "-"}
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>

        <PaginationControls
          basePath="/client"
          limit={limit}
          page={page}
          total={totalCount}
          params={{ start: startParam, end: endParam }}
        />
      </div>
    </div>
  );
}

