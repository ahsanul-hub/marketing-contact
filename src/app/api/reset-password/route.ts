/**
 * Reset Password API - Helper endpoint untuk update password admin
 * 
 * Endpoint ini bisa digunakan untuk update password admin yang sudah ada.
 * Berguna untuk reset password atau fix password yang salah.
 * 
 * POST /api/reset-password
 * Body: { username: "admin", password: "newpassword123" }
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import bcrypt from "bcryptjs";

export async function POST(request: Request) {
  try {
    const body = await request.json();
    const username = String(body.username || "").trim();
    const password = String(body.password || "").trim();

    if (!username || !password) {
      return NextResponse.json(
        { error: "Username and password are required" },
        { status: 400 },
      );
    }

    if (password.length < 6) {
      return NextResponse.json(
        { error: "Password must be at least 6 characters" },
        { status: 400 },
      );
    }

    // Check if admin exists
    const existingAdmin = await prisma.admin.findUnique({
      where: { username },
    });

    if (!existingAdmin) {
      return NextResponse.json(
        { error: "Admin not found" },
        { status: 404 },
      );
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert string hash to Buffer (bytea)
    const passwordBuffer = Buffer.from(hashedPassword, "utf-8");

    // Update password
    const admin = await prisma.admin.update({
      where: { username },
      data: {
        password: passwordBuffer,
      },
      select: {
        id: true,
        username: true,
        role: true,
        isActive: true,
        updatedAt: true,
      },
    });

    return NextResponse.json(
      {
        message: "Password updated successfully",
        admin: {
          username: admin.username,
          role: admin.role,
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
