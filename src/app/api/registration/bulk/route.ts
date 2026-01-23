import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const rawList: unknown = body.phoneNumbers;

    if (!Array.isArray(rawList)) {
      return NextResponse.json(
        { error: "phoneNumbers must be an array" },
        { status: 400 },
      );
    }

    const now = new Date();
    const cleaned = rawList
      .map((v) => String(v || "").trim())
      .filter((v) => v.length > 0);

    if (cleaned.length === 0) {
      return NextResponse.json(
        { error: "No valid phone numbers provided" },
        { status: 400 },
      );
    }

    const uniqueSet = Array.from(new Set(cleaned));

    const result = await prisma.registration.createMany({
      data: uniqueSet.map((phone) => ({
        phoneNumber: phone,
        createdAt: now,
      })),
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

