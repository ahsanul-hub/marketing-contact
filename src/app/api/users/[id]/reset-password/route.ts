/**
 * Reset Password API - Admin Only
 * 
 * Endpoint ini digunakan oleh admin untuk reset password user.
 * Hanya admin yang bisa mengakses endpoint ini.
 * 
 * PUT /api/users/[id]/reset-password
 * Body: { password: "newpassword123" }
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import bcrypt from "bcryptjs";
import { auth } from "@/auth";
import { createActivityLog } from "@/lib/activity-log";

export async function PUT(
  request: Request,
  { params }: { params: { id: string } }
) {
  try {
    // Auth check - pastikan user sudah login dan admin
    const session = await auth();
    if (!session || (session.user as any)?.role !== "admin") {
      return NextResponse.json(
        { error: "Unauthorized" },
        { status: 401 },
      );
    }

    const userId = parseInt(params.id);
    if (isNaN(userId)) {
      return NextResponse.json(
        { error: "Invalid user ID" },
        { status: 400 },
      );
    }

    const body = await request.json();
    const password = String(body.password || "").trim();

    if (!password) {
      return NextResponse.json(
        { error: "Password is required" },
        { status: 400 },
      );
    }

    if (password.length < 6) {
      return NextResponse.json(
        { error: "Password must be at least 6 characters" },
        { status: 400 },
      );
    }

    // Check if user exists
    const existingUser = await prisma.user.findUnique({
      where: { id: userId },
    });

    if (!existingUser) {
      return NextResponse.json(
        { error: "User not found" },
        { status: 404 },
      );
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert string hash to Buffer (bytea)
    const passwordBuffer = Buffer.from(hashedPassword, "utf-8");

    // Update password
    const user = await prisma.user.update({
      where: { id: userId },
      data: {
        password: passwordBuffer,
        updated_at: new Date(),
      },
      select: {
        id: true,
        username: true,
        role: true,
        active: true,
        updated_at: true,
      },
    });

    // Log activity
    if (session?.user?.id) {
      await createActivityLog(
        Number(session.user.id),
        `Reset password for user: ${user.username}`,
      );
    }

    return NextResponse.json(
      {
        message: "Password reset successfully",
        user: {
          id: user.id,
          username: user.username,
          role: user.role,
        },
      },
      { status: 200 },
    );
  } catch (error: any) {
    console.error("Error resetting password", error);
    
    return NextResponse.json(
      { 
        error: "Failed to reset password",
        details: error.message || String(error),
      },
      { status: 500 },
    );
  }
}
