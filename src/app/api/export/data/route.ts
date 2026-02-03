import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { parseDateRangeParams } from "@/lib/pagination";
import { generateExcelBuffer } from "@/lib/excel-template";
import { Prisma } from "@prisma/client";
import dayjs from "dayjs";

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const { startDate, endDate } = parseDateRangeParams({
      start: searchParams.get("start") || "",
      end: searchParams.get("end") || "",
    });

    const searchParam = searchParams.get("search");

    // Build the date filter - only apply if dates are provided
    let dateFilterSql = Prisma.empty;
    if (startDate && endDate) {
      const filterStartDate = dayjs(startDate).startOf("day").toDate();
      const filterEndDate = dayjs(endDate).endOf("day").toDate();
      dateFilterSql = Prisma.sql` AND created_at >= ${filterStartDate} AND created_at <= ${filterEndDate}`;
    }

    // Build the search filter - only apply if search param is provided
    const searchFilterSql = searchParam
      ? Prisma.sql` AND (
          whatsapp ILIKE ${`%${searchParam}%`} OR
          name ILIKE ${`%${searchParam}%`} OR
          nik ILIKE ${`%${searchParam}%`} OR
          owner_name ILIKE ${`%${searchParam}%`}
        )`
      : Prisma.empty;

    const rows = await prisma.$queryRaw<
      { id: bigint; whatsapp: string | null; name: string | null; nik: string | null; owner_name: string | null; created_at: Date | null }[]
    >`
      SELECT id, whatsapp, name, nik, owner_name, created_at
      FROM data
      WHERE 1=1
      ${dateFilterSql}
      ${searchFilterSql}
      ORDER BY created_at DESC
    `;

    const headers = ["Whatsapp", "Name", "NIK", "Owner", "Created At"];
    const data = rows.map((item) => [
      item.whatsapp || "",
      item.name || "",
      item.nik || "",
      item.owner_name || "",
      item.created_at ? dayjs(item.created_at).format("YYYY-MM-DD HH:mm:ss") : "",
    ]);

    const buffer = generateExcelBuffer(headers, data, "Data");

    return new NextResponse(buffer, {
      headers: {
        "Content-Type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "Content-Disposition": `attachment; filename="data-export.xlsx"`,
      },
    });
  } catch (error) {
    console.error("Export data error:", error);
    return NextResponse.json(
      { error: "Failed to export data" },
      { status: 500 }
    );
  }
}