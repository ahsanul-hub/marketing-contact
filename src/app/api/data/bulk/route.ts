import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const rawList: unknown = body.data;

    if (!Array.isArray(rawList)) {
      return NextResponse.json(
        { error: "data must be an array" },
        { status: 400 },
      );
    }

    if (rawList.length === 0) {
      return NextResponse.json({ error: "No data provided" }, { status: 400 });
    }

    const now = new Date();
    const cleaned = rawList.map((item: any) => {
      const whatsapp = String(item.whatsapp || "").trim() || null;
      const name = String(item.name || "").trim() || null;
      const ownerName = String(item.ownerName || item.owner_name || item.clientName || "").trim() || null;

      return {
        whatsapp,
        name,
        ownerName,
        createdAt: now,
      };
    });

    // Insert data directly with ownerName field
    const result = await prisma.data.createMany({
      data: cleaned,
      skipDuplicates: false,
    });

    // Log activity
    const session = await auth();
    if (session?.user?.id && result.count > 0) {
      await createActivityLog(
        Number(session.user.id),
        `Bulk imported ${result.count} data records`,
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
    console.error("Error bulk inserting data", error);
    return NextResponse.json(
      { error: "Failed to insert data" },
      { status: 500 },
    );
  }
}
