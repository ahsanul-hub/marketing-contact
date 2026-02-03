/**
 * Analytics Service
 * 
 * Service ini berisi fungsi-fungsi untuk menghitung data analytics:
 * - Deposit & Profit series untuk chart
 * - Weekly profit bars
 * - Overview metrics (deposit, profit, registrations, contacts)
 * 
 * Semua fungsi menggunakan filter untuk date range dan client filtering.
 */

import { prisma } from "@/lib/prisma";
import dayjs from "dayjs";
import { Prisma } from "@prisma/client";

type TimeFrame = "monthly" | "yearly" | (string & {});

/**
 * Filter untuk analytics queries
 */
export type AnalyticsFilter = {
  startDate?: Date;      // Start date untuk filter
  endDate?: Date;        // End date untuk filter
  clientId?: number;     // Filter by specific client ID (int)
  isOrganic?: boolean;   // Filter untuk organic (tanpa client)
};

export async function getDepositProfitSeries(
  timeFrame: TimeFrame = "monthly",
  filter?: AnalyticsFilter,
) {
  const now = dayjs();
  const start =
    timeFrame === "yearly"
      ? now.subtract(11, "month").startOf("month")
      : now.subtract(29, "day").startOf("day");

  const startBoundary = filter?.startDate ?? start.toDate();
  const endBoundary = filter?.endDate;

  const clientFilterSql =
    filter?.isOrganic === true
      ? Prisma.sql` AND NOT EXISTS (
          SELECT 1 FROM data d
          WHERE d.whatsapp = t.phone_number
        )`
      : filter?.clientId
        ? Prisma.sql` AND t.id_client = ${filter.clientId}`
        : Prisma.empty;

  const dateFilterSql = Prisma.sql`
    AND t.transaction_date >= ${startBoundary}
    ${endBoundary ? Prisma.sql` AND t.transaction_date <= ${endBoundary}` : Prisma.empty}
  `;

  const rows = await prisma.$queryRaw<
    { transaction_date: Date; total_deposit: bigint | null; total_profit: bigint | null }[]
  >`
    SELECT t.transaction_date, t.total_deposit, t.total_profit
    FROM transaction t
    WHERE 1=1
      ${dateFilterSql}
      ${clientFilterSql}
    ORDER BY t.transaction_date ASC
  `;

  if (timeFrame === "yearly") {
    const buckets = buildMonthlyBuckets(start, now);
    rows.forEach((row) => {
      const key = dayjs(row.transaction_date).format("YYYY-MM");
      const bucket = buckets.get(key);
      if (bucket) {
        bucket.deposit += Number(row.total_deposit ?? 0n);
        bucket.profit += Number(row.total_profit ?? 0n);
      }
    });

    const labels = Array.from(buckets.keys());
    return {
      deposit: labels.map((label) => ({
        x: label,
        y: buckets.get(label)?.deposit ?? 0,
      })),
      profit: labels.map((label) => ({
        x: label,
        y: buckets.get(label)?.profit ?? 0,
      })),
    };
  }

  const buckets = buildDailyBuckets(start, now);
  rows.forEach((row) => {
    const key = dayjs(row.transaction_date).format("YYYY-MM-DD");
    const bucket = buckets.get(key);
    if (bucket) {
      bucket.deposit += Number(row.total_deposit ?? 0n);
      bucket.profit += Number(row.total_profit ?? 0n);
    }
  });

  const labels = Array.from(buckets.keys());
  return {
    deposit: labels.map((label) => ({
      x: label.slice(5), // MM-DD
      y: buckets.get(label)?.deposit ?? 0,
    })),
    profit: labels.map((label) => ({
      x: label.slice(5),
      y: buckets.get(label)?.profit ?? 0,
    })),
  };
}

export async function getWeeklyProfitBars(timeFrame?: string, filter?: AnalyticsFilter) {
  const offsetWeeks = timeFrame === "last week" ? 1 : 0;
  const end = dayjs().startOf("day").subtract(7 * offsetWeeks, "day");
  const start = end.subtract(6, "day");

  const clientFilterSql =
    filter?.isOrganic === true
      ? Prisma.sql` AND NOT EXISTS (
          SELECT 1 FROM data d
          WHERE d.whatsapp = t.phone_number
        )`
      : filter?.clientId
        ? Prisma.sql` AND t.id_client = ${filter.clientId}`
        : Prisma.empty;

  const dateFilterSql = Prisma.sql`
    AND t.transaction_date >= ${filter?.startDate ?? start.toDate()}
    AND t.transaction_date <= ${filter?.endDate ?? end.endOf("day").toDate()}
  `;

  const rows = await prisma.$queryRaw<
    { transaction_date: Date; total_profit: bigint | null; total_deposit: bigint | null }[]
  >`
    SELECT t.transaction_date, t.total_profit, t.total_deposit
    FROM transaction t
    WHERE 1=1
      ${dateFilterSql}
      ${clientFilterSql}
    ORDER BY t.transaction_date ASC
  `;

  const buckets = buildDailyBuckets(start, end);
  rows.forEach((row) => {
    const key = dayjs(row.transaction_date).format("YYYY-MM-DD");
    const bucket = buckets.get(key);
    if (bucket) {
      bucket.deposit += Number(row.total_deposit ?? 0n);
      bucket.profit += Number(row.total_profit ?? 0n);
    }
  });

  const labels = Array.from(buckets.keys());
  return {
    deposit: labels.map((label) => ({
      x: dayjs(label).format("ddd"),
      y: buckets.get(label)?.deposit ?? 0,
    })),
    profit: labels.map((label) => ({
      x: dayjs(label).format("ddd"),
      y: buckets.get(label)?.profit ?? 0,
    })),
  };
}

export async function getOverviewMetrics(filter?: AnalyticsFilter) {
  const clientFilterSql =
    filter?.isOrganic === true
      ? Prisma.sql` AND NOT EXISTS (
          SELECT 1 FROM data d
          WHERE d.whatsapp = t.phone_number
        )`
      : filter?.clientId
        ? Prisma.sql` AND t.id_client = ${filter.clientId}`
        : Prisma.empty;

  const dateFilterSql =
    filter?.startDate || filter?.endDate
      ? Prisma.sql`
          AND t.transaction_date >= ${filter.startDate ?? new Date(0)}
          AND t.transaction_date <= ${
            filter.endDate ?? new Date("9999-12-31")
          }
        `
      : Prisma.empty;

  // Get transaction data filtered
  const txAgg = await prisma.$queryRaw<{ sum_deposit: bigint | null; sum_profit: bigint | null }[]>`
    SELECT
      SUM(t.total_deposit)::bigint AS sum_deposit,
      SUM(t.total_profit)::bigint AS sum_profit
    FROM transaction t
    WHERE 1=1
      ${dateFilterSql}
      ${clientFilterSql}
  `;

  // Get registrations with date filter
  const registrationsData = await prisma.$queryRaw<{ count: number }[]>`
    SELECT COUNT(*)::int as count
    FROM registration
    WHERE phone_number IS NOT NULL
      AND created_at >= ${filter?.startDate ?? new Date(0)}
      AND created_at <= ${filter?.endDate ?? new Date("9999-12-31")}
  `;

  // Get unique phone numbers (contacts) with date filter
  const uniquePhones = await prisma.$queryRaw<{ count: bigint }[]>`
    SELECT COUNT(DISTINCT phone_number)::bigint as count
    FROM registration
    WHERE phone_number IS NOT NULL
      AND created_at >= ${filter?.startDate ?? new Date(0)}
      AND created_at <= ${filter?.endDate ?? new Date("9999-12-31")}
  `;

  const txRow = txAgg[0] ?? { sum_deposit: 0n, sum_profit: 0n };

  return {
    deposit: {
      value: Number(txRow.sum_deposit ?? 0n),
      growthRate: 0,
    },
    profit: {
      value: Number(txRow.sum_profit ?? 0n),
      growthRate: 0,
    },
    registrations: {
      value: Number(registrationsData[0]?.count ?? 0n),
      growthRate: 0,
    },
    clients: {
      value: Number(uniquePhones[0]?.count ?? 0n),
      growthRate: 0,
    },
  };
}

function buildDailyBuckets(start: dayjs.Dayjs, end: dayjs.Dayjs) {
  const map = new Map<string, { deposit: number; profit: number }>();
  let cursor = start.startOf("day");
  while (cursor.isBefore(end) || cursor.isSame(end, "day")) {
    map.set(cursor.format("YYYY-MM-DD"), { deposit: 0, profit: 0 });
    cursor = cursor.add(1, "day");
  }
  return map;
}

function buildMonthlyBuckets(start: dayjs.Dayjs, end: dayjs.Dayjs) {
  const map = new Map<string, { deposit: number; profit: number }>();
  let cursor = start.startOf("month");
  while (cursor.isBefore(end) || cursor.isSame(end, "month")) {
    map.set(cursor.format("YYYY-MM"), { deposit: 0, profit: 0 });
    cursor = cursor.add(1, "month");
  }
  return map;
}

export async function getConversionRateData(filter?: AnalyticsFilter) {
  // Count total registrations
  const totalRegistrations = await prisma.$queryRaw<{ count: bigint }[]>`
    SELECT COUNT(DISTINCT id_client)::bigint as count
    FROM registration
    WHERE created_at >= ${filter?.startDate ?? new Date(0)}
      AND created_at <= ${filter?.endDate ?? new Date("9999-12-31")}
  `;

  // Count clients that have transactions
  const clientFilterSql =
    filter?.isOrganic === true
      ? Prisma.sql` AND NOT EXISTS (
          SELECT 1 FROM data d
          WHERE d.whatsapp = t.phone_number
        )`
      : filter?.clientId
        ? Prisma.sql` AND EXISTS (
            SELECT 1 FROM data d
            WHERE d.whatsapp = t.phone_number
              AND d.id_client = ${filter.clientId}
          )`
        : Prisma.empty;

  const clientsWithTransactions = await prisma.$queryRaw<{ count: bigint }[]>`
    SELECT COUNT(DISTINCT id_client)::bigint as count
    FROM transaction t
    WHERE t.transaction_date >= ${filter?.startDate ?? new Date(0)}
      AND t.transaction_date <= ${filter?.endDate ?? new Date("9999-12-31")}
      ${clientFilterSql}
  `;

  const totalReg = Number(totalRegistrations[0]?.count ?? 0);
  const withTx = Number(clientsWithTransactions[0]?.count ?? 0);

  const conversionRate = totalReg > 0 ? (withTx / totalReg) * 100 : 0;

  return {
    conversionRate: parseFloat(conversionRate.toFixed(2)),
    registrations: totalReg,
    transactions: withTx,
  };
}
