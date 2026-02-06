"use client";

import * as XLSX from "xlsx";
import { useState } from "react";
import dayjs from "dayjs";
import { TemplateDownloadButton } from "@/components/template-download-button";

export function TransactionImportForm() {
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

      // Convert to JSON with header row
      const rows: any[] = XLSX.utils.sheet_to_json(sheet, {
        header: 1,
        defval: "",
      });

      if (rows.length < 2) {
        setError("File Excel harus memiliki header dan minimal 1 baris data.");
        return;
      }

      // Get header row (first row)
      const headers = rows[0].map((h: any) =>
        String(h || "").toLowerCase().trim(),
      );

      // Find column indices
      const phoneNumberIdx = headers.findIndex(
        (h: string) =>
          h.includes("phone") ||
          h.includes("nomor") ||
          h.includes("telepon") ||
          h === "phone_number" ||
          h === "phonenumber",
      );
      const dateIdx = headers.findIndex(
        (h: string) =>
          h.includes("date") ||
          h.includes("tanggal") ||
          h.includes("transaction_date") ||
          h === "date",
      );
      const depositIdx = headers.findIndex(
        (h: string) =>
          h.includes("deposit") ||
          h.includes("total_deposit") ||
          h === "deposit",
      );
      const profitIdx = headers.findIndex(
        (h: string) =>
          h.includes("profit") ||
          h.includes("total_profit") ||
          h === "profit",
      );
      const clientIdx = headers.findIndex(
        (h: string) =>
          h.includes("client") ||
          h === "client" ||
          h === "id_client",
      );

      if (dateIdx === -1) {
        setError(
          "Kolom tanggal tidak ditemukan. Pastikan ada kolom 'date' atau 'transaction_date'.",
        );
        return;
      }

      // Process data rows (skip header)
      const transactions = rows.slice(1).map((row: any[]) => {
        const phoneNumber =
          phoneNumberIdx >= 0 ? String(row[phoneNumberIdx] || "").trim() : "";
        const date = row[dateIdx];
        const deposit = depositIdx >= 0 ? row[depositIdx] : 0;
        const profit = profitIdx >= 0 ? row[profitIdx] : 0;
        const client = clientIdx >= 0 ? String(row[clientIdx] || "").trim() : "";

        // Format date to DD-MM-YYYY
        let formattedDate: string;
        if (date instanceof Date) {
          formattedDate = dayjs(date).format("DD-MM-YYYY");
        } else if (typeof date === "string") {
          // Try to parse the date and reformat it
          const parsed = dayjs(date);
          formattedDate = parsed.isValid() ? parsed.format("DD-MM-YYYY") : date;
        } else if (typeof date === "number") {
          // Excel numeric date format
          formattedDate = dayjs(date, "x").format("DD-MM-YYYY");
        } else {
          formattedDate = String(date || "");
        }

        return {
          phoneNumber: phoneNumber || null,
          transactionDate: formattedDate,
          totalDeposit: deposit,
          totalProfit: profit,
          client: client || null,
        };
      });

      if (transactions.length === 0) {
        setError("Tidak ditemukan data transaksi.");
        return;
      }

      const res = await fetch("/api/transaction/bulk", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ transactions }),
      });

      const result = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(result.error || "Gagal import data.");
      }

      setMessage(
        `Berhasil import ${result.inserted} dari ${result.totalSent} transaksi.`,
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
        Bulk import Transaction dari Excel
      </h3>
      <p className="mb-4 text-sm text-neutral-500 dark:text-neutral-300">
        Gunakan file Excel (.xlsx). Format: kolom <strong>date/transaction_date</strong> (format: DD-MM-YYYY) (wajib),{" "}
        <strong>phone_number</strong> (opsional), <strong>total_deposit</strong> (opsional),{" "}
        <strong>total_profit</strong> (opsional).
      </p>

      <div className="mb-4">
        <TemplateDownloadButton type="transaction" />
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
          disabled={uploading || !fileName}
          className="h-10 rounded-md bg-primary px-4 text-sm font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-40"
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