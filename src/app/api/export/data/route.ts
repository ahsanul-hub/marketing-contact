import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { parseDateRangeParams } from "@/lib/pagination";
import { generateExcelBuffer } from "@/lib/excel-template";
import dayjs from "dayjs";

function escapeCSVField(field: string): string {
  if (field.includes(",") || field.includes('"') || field.includes("\n")) {
    return `"${field.replace(/"/g, '""')}"`;
  }
  return field;
}

function generateCSV(headers: string[], data: any[][]): string {
  const headerRow = headers.map(escapeCSVField).join(",");
  const dataRows = data.map((row) =>
    row.map((cell) => escapeCSVField(String(cell || ""))).join(",")
  );
  return [headerRow, ...dataRows].join("\n");
}

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const format = searchParams.get("format") || "xlsx"; // csv or xlsx
    const { startDate, endDate } = parseDateRangeParams({
      start: searchParams.get("start") || "",
      end: searchParams.get("end") || "",
    });

    // Use today as default if no dates provided
    const filterStartDate = startDate || dayjs().startOf("day").toDate();
    const filterEndDate = endDate || dayjs().endOf("day").toDate();

    const where = {
      createdAt: {
        gte: filterStartDate,
        lte: filterEndDate,
      },
    };

    const rows = await prisma.data.findMany({
      orderBy: { createdAt: "desc" },
      where,
      select: {
        whatsapp: true,
        name: true,
        nik: true,
        createdAt: true,
        client: {
          select: {
            name: true,
          },
        },
      },
    });

    const headers = ["Whatsapp", "Nama", "NIK", "Client", "Created At"];
    const data = rows.map((item) => [
      item.whatsapp || "",
      item.name || "",
      item.nik || "",
      item.client?.name || "",
      item.createdAt ? dayjs(item.createdAt).format("YYYY-MM-DD HH:mm:ss") : "",
    ]);

    if (format === "csv") {
      const csvContent = generateCSV(headers, data);
      return new NextResponse(csvContent, {
        headers: {
          "Content-Type": "text/csv; charset=utf-8",
          "Content-Disposition": `attachment; filename="data-export.csv"`,
        },
      });
    } else {
      // Excel format (default)
      const buffer = generateExcelBuffer(headers, data, "Data");
      return new NextResponse(buffer, {
        headers: {
          "Content-Type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
          "Content-Disposition": `attachment; filename="data-export.xlsx"`,
        },
      });
    }
  } catch (error) {
    console.error("Export data error:", error);
    return NextResponse.json(
      { error: "Failed to export data" },
      { status: 500 }
    );
  }
}