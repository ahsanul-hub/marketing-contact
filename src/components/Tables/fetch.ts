import * as logos from "@/assets/logos";

export async function getTopProducts() {
  // Fake delay
  await new Promise((resolve) => setTimeout(resolve, 2000));

  return [
    {
      image: "/images/product/product-01.png",
      name: "Apple Watch Series 7",
      category: "Electronics",
      price: 296,
      sold: 22,
      profit: 45,
    },
    {
      image: "/images/product/product-02.png",
      name: "Macbook Pro M1",
      category: "Electronics",
      price: 546,
      sold: 12,
      profit: 125,
    },
    {
      image: "/images/product/product-03.png",
      name: "Dell Inspiron 15",
      category: "Electronics",
      price: 443,
      sold: 64,
      profit: 247,
    },
    {
      image: "/images/product/product-04.png",
      name: "HP Probook 450",
      category: "Electronics",
      price: 499,
      sold: 72,
      profit: 103,
    },
  ];
}

export async function getInvoiceTableData() {
  // Fake delay
  await new Promise((resolve) => setTimeout(resolve, 1400));

  return [
    {
      name: "Free package",
      price: 0.0,
      date: "2023-01-13T18:00:00.000Z",
      status: "Paid",
    },
    {
      name: "Standard Package",
      price: 59.0,
      date: "2023-01-13T18:00:00.000Z",
      status: "Paid",
    },
    {
      name: "Business Package",
      price: 99.0,
      date: "2023-01-13T18:00:00.000Z",
      status: "Unpaid",
    },
    {
      name: "Standard Package",
      price: 59.0,
      date: "2023-01-13T18:00:00.000Z",
      status: "Pending",
    },
  ];
}

import { prisma } from "@/lib/prisma";
import { Prisma } from "@prisma/client";

export async function getTopProfit(filter?: {
  startDate?: Date;
  endDate?: Date;
  clientId?: number;
  isOrganic?: boolean;
}) {
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

  // Get top profit by phone number, aggregated
  const topProfit = await prisma.$queryRaw<
    {
      phone_number: string | null;
      total_deposit: bigint;
      total_profit: bigint;
    }[]
  >`
    SELECT 
      t.phone_number,
      SUM(t.total_deposit)::bigint as total_deposit,
      SUM(t.total_profit)::bigint as total_profit
    FROM transaction t
    WHERE t.phone_number IS NOT NULL
      ${dateFilterSql}
      ${clientFilterSql}
    GROUP BY t.phone_number
    ORDER BY total_profit DESC
    LIMIT 10
  `;

  return topProfit.map((item) => ({
    phoneNumber: item.phone_number || "-",
    totalDeposit: Number(item.total_deposit),
    totalProfit: Number(item.total_profit),
  }));
}
