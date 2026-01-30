import { prisma } from "@/lib/prisma";
import { NextRequest, NextResponse } from "next/server";

export async function POST(request: NextRequest) {
  try {
    const body = await request.json();
    const { id, phoneNumber, createdAt } = body;

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

    // Build update payload conditionally
    const data: any = {};
    if (typeof phoneNumber === "string") data.phoneNumber = phoneNumber;
    if (createdAt) {
      // attempt to parse createdAt (expect ISO or datetime-local)
      const parsed = new Date(createdAt);
      if (!isNaN(parsed.getTime())) {
        data.createdAt = parsed;
      }
    }

    // If nothing to update, return 400
    if (Object.keys(data).length === 0) {
      return NextResponse.json(
        { message: "Tidak ada field yang diperbarui" },
        { status: 400 }
      );
    }

    // Update registrasi
    const updated = await prisma.registration.update({
      where: { id: registrationId },
      data,
    });

    return NextResponse.json({
      success: true,
      data: {
        ...updated,
        id: updated.id.toString(),
        clientId: updated.clientId.toString(),
      },
      message: "Data berhasil diperbarui",
    });
  } catch (error) {
    console.error("Error updating registration:", error);
    return NextResponse.json(
      { message: "Terjadi kesalahan saat memperbarui data" },
      { status: 500 }
    );
  }
}
