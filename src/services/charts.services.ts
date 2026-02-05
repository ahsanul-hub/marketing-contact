import { prisma } from "@/lib/prisma";
import { Prisma } from "@prisma/client";

export type ClientsFilter = {
  startDate?: Date;
  endDate?: Date;
  clientId?: number;
  isOrganic?: boolean;
};

export async function getTopClientsData(filter?: ClientsFilter) {
  const dateFilterSql =
    filter?.startDate || filter?.endDate
      ? Prisma.sql`
          AND t.transaction_date >= ${filter.startDate ?? new Date(0)}
          AND t.transaction_date <= ${
            filter.endDate ?? new Date("9999-12-31")
          }
        `
      : Prisma.empty;

  const registrationDateFilterSql =
    filter?.startDate || filter?.endDate
      ? Prisma.sql`
          AND r.created_at >= ${filter.startDate ?? new Date(0)}
          AND r.created_at <= ${
            filter.endDate ?? new Date("9999-12-31")
          }
        `
      : Prisma.empty;

  // Get top 10 clients by total deposit and profit from transactions
  const topClients = await prisma.$queryRaw<
    {
      client_id: number | null;
      client_name: string | null;
      total_deposit: bigint;
      total_profit: bigint;
    }[]
  >`
    SELECT 
      c.id as client_id,
      c.name as client_name,
      COALESCE(SUM(t.total_deposit), 0)::bigint as total_deposit,
      COALESCE(SUM(t.total_profit), 0)::bigint as total_profit
    FROM client c
    LEFT JOIN transaction t ON t.id_client = c.id
      ${dateFilterSql}
    GROUP BY c.id, c.name
    ORDER BY total_profit DESC, total_deposit DESC
    LIMIT 10
  `;

  // Get organic (no client) data
  const organicData = await prisma.$queryRaw<{
    total_deposit: bigint;
    total_profit: bigint;
  }[]>`
    SELECT 
      COALESCE(SUM(t.total_deposit), 0)::bigint as total_deposit,
      COALESCE(SUM(t.total_profit), 0)::bigint as total_profit
    FROM transaction t
    WHERE t.phone_number IS NOT NULL
      ${dateFilterSql}
      AND NOT EXISTS (
        SELECT 1 FROM data d WHERE d.whatsapp = t.phone_number
      )
  `;

  const organic = organicData[0] || { total_deposit: 0n, total_profit: 0n };

  // Get conversion rates for each client by counting distinct clients
  // in registration vs transaction tables
  const conversionRates = await prisma.$queryRaw<
    { client_id: number | null; conversion_rate: string }[]
  >`
    SELECT 
      c.id as client_id,
      CASE 
        WHEN reg_count.count = 0 THEN '0'
        ELSE ROUND(
          (COALESCE(tx_count.count, 0)::numeric * 100.0 / 
          reg_count.count),
          2
        )::text
      END as conversion_rate
    FROM client c
    LEFT JOIN (
      SELECT id_client, COUNT(DISTINCT id_client)::bigint as count
      FROM registration
      WHERE created_at >= ${filter?.startDate ?? new Date(0)}
        AND created_at <= ${filter?.endDate ?? new Date("9999-12-31")}
      GROUP BY id_client
    ) reg_count ON c.id = reg_count.id_client
    LEFT JOIN (
      SELECT id_client, COUNT(DISTINCT id_client)::bigint as count
      FROM transaction
      WHERE transaction_date >= ${filter?.startDate ?? new Date(0)}
        AND transaction_date <= ${filter?.endDate ?? new Date("9999-12-31")}
      GROUP BY id_client
    ) tx_count ON c.id = tx_count.id_client
    GROUP BY c.id, reg_count.count, tx_count.count
  `;

  // Get organic conversion rate - registrations without client linked to transactions
  const organicConversionRateData = await prisma.$queryRaw<
    { conversion_rate: string }[]
  >`
    SELECT 
      CASE 
        WHEN reg_count = 0 THEN '0'
        ELSE ROUND(
          (COALESCE(tx_count, 0)::numeric * 100.0 / reg_count),
          2
        )::text
      END as conversion_rate
    FROM (
      SELECT 
        COUNT(DISTINCT r.phone_number)::bigint as reg_count,
        SUM(CASE 
          WHEN t.phone_number IS NOT NULL 
            AND NOT EXISTS (SELECT 1 FROM data d WHERE d.whatsapp = t.phone_number)
          THEN 1 
          ELSE 0
        END)::bigint as tx_count
      FROM registration r
      LEFT JOIN transaction t ON t.phone_number = r.phone_number
        AND t.transaction_date >= ${filter?.startDate ?? new Date(0)}
        AND t.transaction_date <= ${filter?.endDate ?? new Date("9999-12-31")}
      WHERE NOT EXISTS (
        SELECT 1 FROM data d WHERE d.whatsapp = r.phone_number
      )
        AND r.created_at >= ${filter?.startDate ?? new Date(0)}
        AND r.created_at <= ${filter?.endDate ?? new Date("9999-12-31")}
    ) sq
  `;

  const conversionRateMap = new Map<number | null, number>();
  conversionRates.forEach((item) => {
    conversionRateMap.set(item.client_id, parseFloat(item.conversion_rate));
  });
  const organicConversionRate = parseFloat(
    organicConversionRateData[0]?.conversion_rate ?? "0"
  );

  // Combine and sort all data (without Organic)
  const allClients = [
    ...topClients.map((item) => ({
      name: item.client_name || `Client #${item.client_id?.toString()}`,
      totalDeposit: Number(item.total_deposit),
      totalProfit: Number(item.total_profit),
      conversionRate: conversionRateMap.get(item.client_id) ?? 0,
    })),
  ];

  // Sort by profit descending and take top 10
  return allClients
    .sort((a, b) => b.totalProfit - a.totalProfit)
    .slice(0, 10);
}

export async function getDevicesUsedData() {
  const rows = await prisma.$queryRaw<
    { name: string | null; count: bigint }[]
  >`
    SELECT
      c.name,
      COUNT(DISTINCT r.phone_number)::bigint AS count
    FROM registration r
    JOIN client c ON c.id = r.id_client
    GROUP BY c.name
    ORDER BY c.name ASC
  `;

  return rows.map((row) => ({
    name: row.name || "Unknown Client",
    amount: Number(row.count),
  }));
}
