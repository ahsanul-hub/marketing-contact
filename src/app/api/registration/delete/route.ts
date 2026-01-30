import { prisma } from "@/lib/prisma";
import { NextRequest, NextResponse } from "next/server";

export async function DELETE(request: NextRequest) {
  try {
    const body = await request.json();
    const { id } = body;

    if (!id) {
      return NextResponse.json(
        { message: "ID registrasi diperlukan" },
        { status: 400 }
      );
    }

    const registrationId = BigInt(id);

    // Cek apakah registrasi ada
    const existing = await prisma.registration.findUnique({
      where: { id: registrationId },
    });

    if (!existing) {
      return NextResponse.json(
        { message: "Data registrasi tidak ditemukan" },
        { status: 404 }
      );
    }

    // Delete registrasi
    await prisma.registration.delete({
      where: { id: registrationId },
    });

    return NextResponse.json({
      success: true,
      message: "Data berhasil dihapus",
    });
  } catch (error) {
    console.error("Error deleting registration:", error);
    return NextResponse.json(
      { message: "Terjadi kesalahan saat menghapus data" },
      { status: 500 }
    );
  }
}
