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
    console.log("POST /api/clients: Starting request");
    
    const session = await auth();
    console.log("Session:", session ? "Authenticated" : "Not authenticated");
    
    const body = await request.json();
    console.log("Request body:", body);
    
    const name = String(body.name || "").trim();
    console.log("Parsed name:", name);
    
    if (!name) {
      console.log("Validation failed: Name is required");
      return NextResponse.json(
        { error: "Name is required" },
        { status: 400 },
      );
    }
    
    console.log("Creating client with name:", name);
    const client = await prisma.client.create({
      data: {
        name,
        createdAt: new Date(),
      },
    });
    console.log("Client created:", client);

    // Log activity
    if (session?.user?.id) {
      const userId = Number(session.user.id);
      console.log("Logging activity for user:", userId, "type:", typeof userId);
      if (!isNaN(userId) && userId > 0) {
        await createActivityLog(
          userId,
          `Created client: ${name}`,
        );
      } else {
        console.log("Invalid user ID, skipping activity log");
      }
    } else {
      console.log("No session or user ID, skipping activity log");
    }

    // Convert BigInt to string for JSON serialization
    const clientResponse = {
      ...client,
      id: client.id.toString(),
    };

    return NextResponse.json(clientResponse, { status: 201 });
  } catch (error) {
    console.error("Error creating client:", error);
    return NextResponse.json(
      { error: "Failed to create client" },
      { status: 500 },
    );
  }
}
