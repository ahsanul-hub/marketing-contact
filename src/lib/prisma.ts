/**
 * Prisma Client Singleton
 * 
 * File ini membuat instance Prisma Client yang digunakan di seluruh aplikasi.
 * Menggunakan singleton pattern untuk menghindari multiple instances di development.
 * 
 * IMPORTANT: Pastikan DATABASE_URL sudah di-set di .env.local
 * Format: postgresql://user:password@host:port/database
 * 
 * Usage:
 * import { prisma } from "@/lib/prisma";
 * const users = await prisma.user.findMany();
 */

import { PrismaClient } from "@prisma/client";

// Singleton pattern untuk development - hindari multiple instances saat hot-reload
const globalForPrisma = globalThis as unknown as {
  prisma?: PrismaClient;
};

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    // Log queries, errors, dan warnings di development
    log: ["query", "error", "warn"],
  });

// Simpan instance di global untuk reuse di development
if (process.env.NODE_ENV !== "production") {
  globalForPrisma.prisma = prisma;
}

