/**
 * Setup Admin API Endpoint
 * 
 * Endpoint ini digunakan untuk membuat admin pertama kali.
 * Hanya bisa digunakan jika belum ada admin di database.
 * 
 * Usage:
 * POST /api/setup
 * Body: { username: "admin", password: "password123" }
 * 
 * Setelah admin dibuat, gunakan /admin/users untuk menambah admin lainnya.
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import bcrypt from "bcryptjs";

export async function POST(request: Request) {
  try {
    // Check if admin already exists
    const existingUser = await prisma.user.findFirst({
      where: { role: "admin" },
    });

    if (existingUser) {
      return NextResponse.json(
        { error: "Admin already exists. Please use /admin/users to add more admins." },
        { status: 400 },
      );
    }

    const body = await request.json();
    const username = String(body.username || "").trim();
    const password = String(body.password || "").trim();

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

    if (password.length < 6) {
      return NextResponse.json(
        { error: "Password must be at least 6 characters" },
        { status: 400 },
      );
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert string hash to Buffer (bytea)
    const passwordBuffer = Buffer.from(hashedPassword, "utf-8");

    // Create admin user
    const user = await prisma.user.create({
      data: {
        username,
        password: passwordBuffer,
        role: "admin",
        active: true,
        updated_at: new Date(),
      },
      select: {
        id: true,
        username: true,
        role: true,
      },
    });

    return NextResponse.json(
      {
        message: "Admin created successfully",
        user: user,
      },
      { status: 201 },
    );
  } catch (error: any) {
    console.error("Error setting up admin", error);
    console.error("Error details:", JSON.stringify(error, null, 2));
    
    if (error.code === "P2002") {
      return NextResponse.json(
        { error: "Username already exists" },
        { status: 400 },
      );
    }
    
    return NextResponse.json(
      { 
        error: "Failed to setup admin",
        details: error.message || String(error),
        code: error.code || "UNKNOWN"
      },
      { status: 500 },
    );
  }
}
