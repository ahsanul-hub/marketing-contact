"use client";

import * as XLSX from "xlsx";
import { useState } from "react";
import { TemplateDownloadButton } from "@/components/template-download-button";

export function RegistrationImportForm() {
  const [fileName, setFileName] = useState<string | null>(null);
  const [uploading, setUploading] = useState(false);
  const [message, setMessage] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleFileChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) {
      setFileName(null);
      return;
    }
    setFileName(file.name);
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setMessage(null);
    setError(null);

    const input = e.currentTarget.elements.namedItem(
      "file",
    ) as HTMLInputElement | null;
    const file = input?.files?.[0];

    if (!file) {
      setError("Pilih file Excel terlebih dahulu");
      return;
    }

    try {
      setUploading(true);

      const data = await file.arrayBuffer();
      const workbook = XLSX.read(data, { type: "array" });
      const sheetName = workbook.SheetNames[0];
      const sheet = workbook.Sheets[sheetName];

      const rows: any[][] = XLSX.utils.sheet_to_json(sheet, {
        header: 1,
        defval: "",
      });

      // Ambil semua nilai di kolom pertama (A) sebagai phone_number
      const phones = rows
        .map((row) => String(row[0] ?? "").trim())
        .filter((v) => v.length > 0);

      if (phones.length === 0) {
        setError("Tidak ditemukan nomor telepon di kolom pertama file.");
        return;
      }

      const res = await fetch("/api/registration/bulk", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ phoneNumbers: phones }),
      });

      const result = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(result.error || "Gagal import data.");
      }

      setMessage(
        `Berhasil insert ${result.inserted} dari ${result.totalSent} nomor (duplikat di-skip).`,
      );
    } catch (err: any) {
      console.error(err);
      setError(err.message || "Terjadi kesalahan saat import.");
    } finally {
      setUploading(false);
    }
  };

  return (
    <div className="rounded-[10px] border border-dashed border-stroke bg-white p-4 shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card sm:p-5">
      <h3 className="mb-2 text-base font-semibold text-dark dark:text-white">
        Bulk import Registration dari Excel
      </h3>
      <p className="mb-4 text-sm text-neutral-500 dark:text-neutral-300">
        Gunakan file Excel (.xlsx). Sistem akan membaca nomor dari kolom
        pertama (kolom A).
      </p>

      <div className="mb-4">
        <TemplateDownloadButton type="registration" />
      </div>

      <form onSubmit={handleSubmit} className="flex flex-wrap items-end gap-3">
        <div>
          <input
            type="file"
            name="file"
            accept=".xlsx,.xls"
            onChange={handleFileChange}
            className="block w-64 text-sm text-neutral-700 file:mr-4 file:rounded-md file:border-0 file:bg-primary file:px-3 file:py-2 file:text-sm file:font-medium file:text-white hover:file:bg-opacity-90 dark:text-neutral-200"
          />
          {fileName && (
            <p className="mt-1 text-xs text-neutral-500 dark:text-neutral-400">
              {fileName}
            </p>
          )}
        </div>

        <button
          type="submit"
          disabled={uploading}
          className="h-10 rounded-md bg-primary px-4 text-sm font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-60"
        >
          {uploading ? "Mengupload..." : "Import"}
        </button>
      </form>

      {message && (
        <p className="mt-3 text-sm text-emerald-600 dark:text-emerald-400">
          {message}
        </p>
      )}

      {error && (
        <p className="mt-3 text-sm text-red-500 dark:text-red-400">{error}</p>
      )}
    </div>
  );
}

