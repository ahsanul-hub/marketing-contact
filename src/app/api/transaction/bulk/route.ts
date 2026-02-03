import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import dayjs from "dayjs";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const rawList: unknown = body.transactions;

    if (!Array.isArray(rawList)) {
      return NextResponse.json(
        { error: "transactions must be an array" },
        { status: 400 },
      );
    }

    if (rawList.length === 0) {
      return NextResponse.json(
        { error: "No transactions provided" },
        { status: 400 },
      );
    }

    const cleaned = rawList
      .map((item: any) => {
        const phoneNumber = String(item.phoneNumber || item.phone_number || "").trim();
        const transactionDate = item.transactionDate || item.transaction_date || item.date;
        const totalDeposit = item.totalDeposit || item.total_deposit || item.deposit || 0;
        const totalProfit = item.totalProfit || item.total_profit || item.profit || 0;
        const client = item.client || item.client_name || item.id_client;

        // Parse date
        let parsedDate: Date;
        if (transactionDate instanceof Date) {
          parsedDate = transactionDate;
        } else if (typeof transactionDate === "string") {
          const d = dayjs(transactionDate);
          parsedDate = d.isValid() ? d.toDate() : new Date();
        } else {
          parsedDate = new Date();
        }

        // Parse numbers
        const deposit = BigInt(Math.max(0, Number(totalDeposit) || 0));
        const profit = BigInt(Math.max(0, Number(totalProfit) || 0));

        return {
          phoneNumber: phoneNumber || null,
          transactionDate: parsedDate,
          totalDeposit: deposit,
          totalProfit: profit,
          client,
        };
      })
      .filter((item) => item.transactionDate);

    if (cleaned.length === 0) {
      return NextResponse.json(
        { error: "No valid transactions provided" },
        { status: 400 },
      );
    }

    // Get all clients for mapping
    const clients = await prisma.client.findMany({
      select: { id: true, name: true },
    });
    const clientMap = new Map(clients.map(c => [c.name?.toLowerCase(), c.id]));

    // Resolve clientId
    const data = cleaned.map(item => {
      let clientId: number | null = null;
      if (item.client) {
        if (typeof item.client === 'string') {
          const lowerName = item.client.toLowerCase();
          clientId = clientMap.get(lowerName) || null;
        } else {
          clientId = (item.client);
        }
      }
      const record: any = {
        phoneNumber: item.phoneNumber,
        transactionDate: item.transactionDate,
        totalDeposit: item.totalDeposit,
        totalProfit: item.totalProfit,
      };
      if (clientId !== null) {
        record.clientId = clientId;
      }
      return record;
    });

    const result = await prisma.transaction.createMany({
      data,
      skipDuplicates: false,
    });

    // Log activity
    const session = await auth();
    if (session?.user?.id && result.count > 0) {
      await createActivityLog(
        Number(session.user.id),
        `Bulk imported ${result.count} transactions`,
      );
    }

    return NextResponse.json(
      {
        inserted: result.count,
        totalSent: cleaned.length,
      },
      { status: 201 },
    );
  } catch (error) {
    console.error("Error bulk inserting transactions", error);
    return NextResponse.json(
      { error: "Failed to insert transactions" },
      { status: 500 },
    );
  }
}