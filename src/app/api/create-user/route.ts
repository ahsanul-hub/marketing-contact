/**
 * Create Admin API - Helper endpoint untuk membuat admin
 * 
 * Endpoint ini bisa digunakan untuk membuat admin tanpa perlu login.
 * Berguna untuk setup pertama kali atau testing.
 * 
 * POST /api/create-user
 * Body: { username: "admin", password: "password123", role: "admin" }
 */

import { NextResponse } from "next/server";
import { prisma } from "@/lib/prisma";
import bcrypt from "bcryptjs";

export async function POST(request: Request) {
  try {
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

    if (password.length < 6) {
      return NextResponse.json(
        { error: "Password must be at least 6 characters" },
        { status: 400 },
      );
    }

    // Check if admin already exists
    const existingAdmin = await prisma.admin.findUnique({
      where: { username },
    });

    if (existingAdmin) {
      return NextResponse.json(
        { error: "Admin with this username already exists" },
        { status: 400 },
      );
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert string hash to Buffer (bytea)
    const passwordBuffer = Buffer.from(hashedPassword, "utf-8");

    // Create admin
    const admin = await prisma.admin.create({
      data: {
        username,
        password: passwordBuffer,
        role: role === "admin" ? "admin" : "user",
        isActive: true,
      },
      select: {
        id: true,
        username: true,
        role: true,
        isActive: true,
        createdAt: true,
      },
    });

    return NextResponse.json(
      {
        message: "Admin created successfully",
        admin: admin,
      },
      { status: 201 },
    );
  } catch (error: any) {
    console.error("Error creating admin", error);
    
    if (error.code === "P2002") {
      return NextResponse.json(
        { error: "Username already exists" },
        { status: 400 },
      );
    }
    
    return NextResponse.json(
      { 
        error: "Failed to create admin",
        details: error.message || String(error),
      },
      { status: 500 },
    );
  }
}
