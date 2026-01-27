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
        user: {
          select: {
            id: true,
            username: true,
          },
        },
      },
    });

  const safeLogs = logs.map(log => ({
    ...log,
    id: log.id.toString(),
    user_id: log.userId?.toString(),
  }));
  return NextResponse.json(safeLogs);

  } catch (error) {
    console.error("Error fetching activity logs", error);
    return NextResponse.json(
      { error: "Failed to fetch activity logs" },
      { status: 500 },
    );
  }
}