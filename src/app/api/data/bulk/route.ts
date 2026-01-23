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
      return NextResponse.json(
        { error: "No data provided" },
        { status: 400 },
      );
    }

    const now = new Date();
    const cleaned = rawList.map((item: any) => {
      const whatsapp = String(item.whatsapp || "").trim() || null;
      const name = String(item.name || "").trim() || null;
      const nik = String(item.nik || "").trim() || null;
      const clientName = String(item.client || item.clientName || "").trim();

      return {
        whatsapp,
        name,
        nik,
        clientName: clientName || null,
        createdAt: now,
      };
    });

    // Get or create clients
    const clientMap = new Map<string, bigint>();
    const uniqueClientNames = Array.from(
      new Set(cleaned.map((item) => item.clientName).filter(Boolean)),
    );

    if (uniqueClientNames.length > 0) {
      // Fetch existing clients
      const existingClients = await prisma.client.findMany({
        where: {
          name: { in: uniqueClientNames },
        },
      });

      existingClients.forEach((client) => {
        if (client.name) {
          clientMap.set(client.name, client.id);
        }
      });

      // Create missing clients
      const missingClientNames = uniqueClientNames.filter(
        (name) => !clientMap.has(name),
      );

      if (missingClientNames.length > 0) {
        const newClients = await prisma.client.createMany({
          data: missingClientNames.map((name) => ({
            name,
            createdAt: now,
          })),
        });

        // Fetch the newly created clients to get their IDs
        const createdClients = await prisma.client.findMany({
          where: {
            name: { in: missingClientNames },
          },
        });

        createdClients.forEach((client) => {
          if (client.name) {
            clientMap.set(client.name, client.id);
          }
        });
      }
    }

    // Prepare data for insertion
    const dataToInsert = cleaned.map((item) => ({
      whatsapp: item.whatsapp,
      name: item.name,
      nik: item.nik,
      createdAt: item.createdAt,
      clientId: item.clientName ? clientMap.get(item.clientName) || null : null,
    }));

    const result = await prisma.data.createMany({
      data: dataToInsert,
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
