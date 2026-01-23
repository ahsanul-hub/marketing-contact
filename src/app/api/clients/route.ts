/**
 * Client API Routes
 * 
 * GET /api/clients - List semua clients
 * POST /api/clients - Create client baru (require authentication)
 * 
 * Activity log akan dicatat saat create client.
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function GET() {
  try {
    const clients = await prisma.client.findMany({
      orderBy: [{ createdAt: "desc" }, { id: "desc" }],
    });

    return NextResponse.json(clients);
  } catch (error) {
    console.error("Error fetching clients", error);
    return NextResponse.json(
      { error: "Failed to fetch clients" },
      { status: 500 },
    );
  }
}

export async function POST(request: Request) {
  try {
    const session = await auth();
    const body = await request.json();
    const name = String(body.name || "").trim();
    
    if (!name) {
      return NextResponse.json(
        { error: "Name is required" },
        { status: 400 },
      );
    }
    
    const client = await prisma.client.create({
      data: {
        name,
        createdAt: new Date(),
      },
    });

    // Log activity
    if (session?.user?.id) {
      await createActivityLog(
        Number(session.user.id),
        `Created client: ${name}`,
      );
    }

    return NextResponse.json(client, { status: 201 });
  } catch (error) {
    console.error("Error creating client", error);
    return NextResponse.json(
      { error: "Failed to create client" },
      { status: 500 },
    );
  }
}
