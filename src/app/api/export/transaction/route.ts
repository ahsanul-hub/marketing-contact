import { NextRequest, NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { parseDateRangeParams } from "@/lib/pagination";
import { generateExcelBuffer } from "@/lib/excel-template";
import dayjs from "dayjs";

export async function GET(request: NextRequest) {
  try {
    const { searchParams } = new URL(request.url);
    const { startDate, endDate } = parseDateRangeParams({
      start: searchParams.get("start") ?? "",
      end: searchParams.get("end") ?? "",
    });

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
    };

    const transactions = await prisma.transaction.findMany({
      orderBy: { transactionDate: "desc" },
      where,
      select: {
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

    const headers = ["Transaction Date", "Phone Number", "Total Deposit", "Total Profit", "Client"];
    const data = transactions.map((item) => [
      dayjs(item.transactionDate).format("YYYY-MM-DD HH:mm:ss"),
      item.phoneNumber || "",
      item.totalDeposit ? Number(item.totalDeposit) : 0,
      item.totalProfit ? Number(item.totalProfit) : 0,
      item.client?.name || "",
    ]);

    const buffer = generateExcelBuffer(headers, data, "Transaction");

    return new NextResponse(buffer, {
      headers: {
        "Content-Type": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
        "Content-Disposition": `attachment; filename="transaction-export.xlsx"`,
      },
    });
  } catch (error) {
    console.error("Export transaction error:", error);
    return NextResponse.json(
      { error: "Failed to export data" },
      { status: 500 }
    );
  }
}