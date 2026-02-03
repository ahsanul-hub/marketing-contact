import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { parseDateRangeParams } from "@/lib/pagination";
import { generateExcelBuffer } from "@/lib/excel-template";
import { AnalyticsFilter } from "@/services/analytics";
import dayjs from "dayjs";

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const { startDate, endDate } = parseDateRangeParams({
      start: searchParams.get("start") ?? "",
      end: searchParams.get("end") ?? "",
    });

    const clientIdParam = searchParams.get("client_id");
    const isOrganic = clientIdParam === "organic";
    const clientId =
      clientIdParam && clientIdParam !== "organic" ? Number(clientIdParam) : undefined;

    const filter: AnalyticsFilter = {
      startDate,
      endDate,
      clientId,
      isOrganic,
    };

    // Get top profit data with filter
    const clientFilterSql =
      filter?.isOrganic === true
        ? ` AND NOT EXISTS (
            SELECT 1 FROM data d
            WHERE d.whatsapp = t.phone_number
          )`
        : filter?.clientId
          ? ` AND t.id_client = ${filter.clientId}`
          : "";

    const dateFilterSql =
      filter?.startDate || filter?.endDate
        ? ` AND t.transaction_date >= '${filter.startDate?.toISOString() ?? '1970-01-01'}'
            AND t.transaction_date <= '${filter.endDate?.toISOString() ?? '9999-12-31'}'`
        : "";

    const topProfit = await prisma.$queryRaw<
      {
        phone_number: string | null;
        total_deposit: bigint;
        total_profit: bigint;
      }[]
    >`
      SELECT
        phone_number,
        SUM(total_deposit)::bigint as total_deposit,
        SUM(total_profit)::bigint as total_profit
      FROM transaction t
      WHERE phone_number IS NOT NULL
        ${dateFilterSql}
        ${clientFilterSql}
      GROUP BY phone_number
      ORDER BY total_profit DESC
      LIMIT 10
    `;

    const headers = ["Phone Number", "Total Deposit", "Total Profit"];
    const data = topProfit.map((item) => [
      item.phone_number || "",
      Number(item.total_deposit),
      Number(item.total_profit),
    ]);

    const buffer = generateExcelBuffer(headers, data, "Top Profit");

    return new NextResponse(buffer, {
      headers: {
        "Content-Type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "Content-Disposition": `attachment; filename="top-profit-export.xlsx"`,
      },
    });
  } catch (error) {
    console.error("Export home error:", error);
    return NextResponse.json(
      { error: "Failed to export data" },
      { status: 500 }
    );
  }
}