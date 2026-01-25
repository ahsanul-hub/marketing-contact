/**
 * Register API - Endpoint untuk mendaftar user baru
 * 
 * Endpoint ini digunakan untuk registrasi user baru dengan default role "client".
 * 
 * POST /api/register
 * Body: { username: "user123", password: "password123" }
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
        { error: "Username dan password wajib diisi" },
        { status: 400 },
      );
    }

    if (username.length > 30) {
      return NextResponse.json(
        { error: "Username maksimal 30 karakter" },
        { status: 400 },
      );
    }

    if (username.length < 3) {
      return NextResponse.json(
        { error: "Username minimal 3 karakter" },
        { status: 400 },
      );
    }

    if (password.length < 6) {
      return NextResponse.json(
        { error: "Password minimal 6 karakter" },
        { status: 400 },
      );
    }

    // Check if user already exists
    const existingUser = await prisma.user.findUnique({
      where: { username },
    });

    if (existingUser) {
      return NextResponse.json(
        { error: "Username sudah digunakan" },
        { status: 400 },
      );
    }

    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert string hash to Buffer (bytea)
    const passwordBuffer = Buffer.from(hashedPassword, "utf-8");

    // Create user dengan default role "client"
    const user = await prisma.user.create({
      data: {
        username,
        password: passwordBuffer,
        role: "client", // Default role untuk user yang daftar
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

    return NextResponse.json(
      {
        message: "Registrasi berhasil",
        user: user,
      },
      { status: 201 },
    );
  } catch (error) {
    console.error("Error registering user", error);
    return NextResponse.json(
      { error: "Gagal melakukan registrasi" },
      { status: 500 },
    );
  }
}

