/**
 * Activity Log Utility
 * 
 * Fungsi untuk mencatat aktivitas user (insert, update, delete) ke database.
 * Digunakan untuk tracking siapa yang melakukan apa dan kapan.
 * 
 * @param userId - ID user yang melakukan action
 * @param action - Action type: "INSERT", "UPDATE", "DELETE"
 * @param details - Optional: Detail tambahan (JSON string atau text)
 * 
 * Note: Function ini tidak throw error karena activity log failure
 * tidak boleh mengganggu operasi utama.
 */

import { prisma } from "@/lib/prisma";

export async function createActivityLog(
  userId: number,
  action: string,
  details?: string,
) {
  try {
    await prisma.activityLog.create({
      data: {
        userId,
        action,
        details: details || null,
      },
    });
  } catch (error) {
    console.error("Failed to create activity log:", error);
    // Don't throw - activity log failure shouldn't break the main operation
  }
}
