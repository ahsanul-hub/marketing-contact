import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { auth } from "@/auth";

export async function GET() {
  try {
    const session = await auth();
    if (!session) {
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 },
      );
    }

    const logs = await prisma.activityLog.findMany({
      take: 20,
      orderBy: { createdAt: "desc" },
      include: {
        admin: {
          select: {
            id: true,
            username: true,
          },
        },
      },
    });

    return NextResponse.json(logs);
  } catch (error) {
    console.error("Error fetching activity logs", error);
    return NextResponse.json(
      { error: "Failed to fetch activity logs" },
      { status: 500 },
    );
  }
}
