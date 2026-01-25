import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import bcrypt from "bcryptjs";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function GET() {
  try {
    const session = await auth();
    if (!session || (session.user as any)?.role !== "admin") {
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 },
      );
    }

    const users = await prisma.user.findMany({
      select: {
        id: true,
        username: true,
        role: true,
        active: true,
        created_at: true,
      },
      orderBy: { created_at: "desc" },
    });

    return NextResponse.json(users);
  } catch (error) {
    console.error("Error fetching admins", error);
    return NextResponse.json(
      { error: "Failed to fetch admins" },
      { status: 500 },
    );
  }
}

export async function POST(request: Request) {
  try {
    const session = await auth();
    if (!session || (session.user as any)?.role !== "admin") {
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 },
      );
    }

    const body = await request.json();
    const username = String(body.username || "").trim();
    const password = String(body.password || "").trim();
    const role = String(body.role || "user").trim();

    if (!username || !password) {
      return NextResponse.json(
        { error: "Username and password are required" },
        { status: 400 },
      );
    }

    if (username.length > 30) {
      return NextResponse.json(
        { error: "Username must be 30 characters or less" },
        { status: 400 },
      );
    }

    // Check if user already exists
    const existingUser = await prisma.user.findUnique({
      where: { username },
    });

    if (existingUser) {
      return NextResponse.json(
        { error: "User with this username already exists" },
        { status: 400 },
      );
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert string hash to Buffer (bytea)
    const passwordBuffer = Buffer.from(hashedPassword, "utf-8");

    // Create user
    const user = await prisma.user.create({
      data: {
        username,
        password: passwordBuffer,
        role: role === "admin" ? "admin" : role === "client" ? "client" : "user",
        active: true,
        updated_at: new Date(),
      },
      select: {
        id: true,
        username: true,
        role: true,
        active: true,
        created_at: true,
      },
    });

    // Log activity
    if (session?.user?.id) {
      await createActivityLog(
        Number(session.user.id),
        `Created ${role} user: ${username}`,
      );
    }

    return NextResponse.json(user, { status: 201 });
  } catch (error) {
    console.error("Error creating admin", error);
    return NextResponse.json(
      { error: "Failed to create admin" },
      { status: 500 },
    );
  }
}
