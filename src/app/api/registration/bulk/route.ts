import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const rawList: unknown = body.registrations || body.phoneNumbers;

    if (!Array.isArray(rawList)) {
      return NextResponse.json(
        { error: "registrations must be an array" },
        { status: 400 },
      );
    }

    // Get all clients for mapping
    const clients = await prisma.client.findMany({
      select: { id: true, name: true },
    });
    const clientMap = new Map(clients.map(c => [c.name?.toLowerCase(), c.id]));

    const cleaned = rawList
      .map((item: any) => {
        if (typeof item === 'string') {
          return { phoneNumber: String(item || "").trim(), client: null, createdAt: null };
        } else {
          return {
            phoneNumber: String(item.phoneNumber || item.phone_number || "").trim(),
            client: item.client || item.client_name || item.id_client || null,
            createdAt: item.createdAt || item.created_at || null,
          };
        }
      })
      .filter((v) => v.phoneNumber.length > 0);

    if (cleaned.length === 0) {
      return NextResponse.json(
        { error: "No valid registrations provided" },
        { status: 400 },
      );
    }

    // Resolve clientId and parse createdAt
    const data = cleaned.map(item => {
      let clientId: bigint | null = null;
      if (item.client) {
        if (typeof item.client === 'string') {
          const lowerName = item.client.toLowerCase();
          clientId = clientMap.get(lowerName) || null;
        } else {
          clientId = BigInt(item.client);
        }
      }

      // Parse createdAt from the provided value
      let createdAt = new Date();
      if (item.createdAt) {
        const parsedDate = new Date(item.createdAt);
        if (!isNaN(parsedDate.getTime())) {
          createdAt = parsedDate;
        }
      }

      return {
        phoneNumber: item.phoneNumber,
        createdAt,
        clientId,
      };
    });

    const uniqueSet = Array.from(new Set(data.map(d => `${d.phoneNumber}-${d.clientId}`)))
      .map(key => {
        const [phone, clientId] = key.split('-');
        return data.find(d => d.phoneNumber === phone && String(d.clientId) === clientId)!;
      })
      .filter((item) => item.clientId !== null) as Array<{ phoneNumber: string; createdAt: Date; clientId: bigint }>;

    const result = await prisma.registration.createMany({
      data: uniqueSet,
      skipDuplicates: true,
    });

    // Log activity
    const session = await auth();
    if (session?.user?.id && result.count > 0) {
      await createActivityLog(
        Number(session.user.id),
        `Bulk imported ${result.count} registrations`,
      );
    }

    return NextResponse.json(
      {
        inserted: result.count,
        totalSent: uniqueSet.length,
      },
      { status: 201 },
    );
  } catch (error) {
    console.error("Error bulk inserting registration", error);
    return NextResponse.json(
      { error: "Failed to insert registrations" },
      { status: 500 },
    );
  }
}
