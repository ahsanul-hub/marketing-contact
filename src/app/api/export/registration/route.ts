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
        start: searchParams.get("start") ?? "",
        end: searchParams.get("end") ?? "",
    });

    const clientIdParam = searchParams.get("client_id");
    const isOrganic = clientIdParam === "organic";
    const clientId =
      clientIdParam && clientIdParam !== "organic" ? BigInt(clientIdParam) : undefined;

    // Use today as default if no dates provided
    const filterStartDate = startDate || dayjs().startOf("day").toDate();
    const filterEndDate = endDate || dayjs().endOf("day").toDate();

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

    const registrations = await prisma.$queryRaw<
      { phone_number: string | null; created_at: Date | null }[]
    >`
      SELECT r.phone_number, r.created_at
      FROM registration r
      WHERE 1=1
        ${dateFilterSql}
        ${typeFilterSql}
      ORDER BY r.created_at DESC NULLS LAST, r.id DESC
    `;

    const headers = ["Phone Number", "Created At"];
    const data = registrations.map((item) => [
      item.phone_number || "",
      item.created_at ? dayjs(item.created_at).format("YYYY-MM-DD HH:mm:ss") : "",
    ]);

    const buffer = generateExcelBuffer(headers, data, "Registration");

    return new NextResponse(buffer, {
      headers: {
        "Content-Type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "Content-Disposition": `attachment; filename="registration-export.xlsx"`,
      },
    });
  } catch (error) {
    console.error("Export registration error:", error);
    return NextResponse.json(
      { error: "Failed to export data" },
      { status: 500 }
    );
  }
}