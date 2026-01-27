"use client";

import * as XLSX from "xlsx";
import { useState } from "react";
import { TemplateDownloadButton } from "@/components/template-download-button";

export function DataImportForm() {
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

  const parseCSV = (text: string): any[][] => {
    const lines = text.split(/\r?\n/).filter((line) => line.trim());
    return lines.map((line) => {
      const result: string[] = [];
      let current = "";
      let inQuotes = false;

      for (let i = 0; i < line.length; i++) {
        const char = line[i];
        const nextChar = line[i + 1];

        if (char === '"') {
          if (inQuotes && nextChar === '"') {
            current += '"';
            i++; // Skip next quote
          } else {
            inQuotes = !inQuotes;
          }
        } else if (char === "," && !inQuotes) {
          result.push(current.trim());
          current = "";
        } else {
          current += char;
        }
      }
      result.push(current.trim());
      return result;
    });
  };

  const processFileData = async (file: File): Promise<any[][]> => {
    const fileExtension = file.name.split(".").pop()?.toLowerCase();

    if (fileExtension === "csv") {
      const text = await file.text();
      const rows = parseCSV(text);
      return rows;
    } else {
      // Excel file
      const data = await file.arrayBuffer();
      const workbook = XLSX.read(data, { type: "array" });
      const sheetName = workbook.SheetNames[0];
      const sheet = workbook.Sheets[sheetName];

      const rows: any[] = XLSX.utils.sheet_to_json(sheet, {
        header: 1,
        defval: "",
      });
      return rows;
    }
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
      setError("Pilih file CSV atau Excel terlebih dahulu");
      return;
    }

    try {
      setUploading(true);

      const rows = await processFileData(file);

      if (rows.length < 2) {
        setError("File harus memiliki header dan minimal 1 baris data.");
        return;
      }

      // Get header row (first row)
      const headers = rows[0].map((h: any) =>
        String(h || "").toLowerCase().trim(),
      );

      // Find column indices - support both Indonesian and English
      const whatsappIdx = headers.findIndex(
        (h: string) =>
          h.includes("whatsapp") ||
          h.includes("wa") ||
          h === "whatsapp",
      );
      const nameIdx = headers.findIndex(
        (h: string) =>
          h.includes("name") ||
          h.includes("nama") ||
          h === "name",
      );
      const nikIdx = headers.findIndex(
        (h: string) =>
          h.includes("nik") ||
          h === "nik",
      );
      const clientIdx = headers.findIndex(
        (h: string) =>
          h.includes("client") ||
          h.includes("klien") ||
          h === "client",
      );

      // If no headers found, assume first 3 columns are Whatsapp, Nama, NIK (based on user's format)
      const hasHeaders = whatsappIdx >= 0 || nameIdx >= 0 || nikIdx >= 0;
      
      let dataRows: any[];
      
      if (!hasHeaders && rows[0].length >= 3) {
        // Assume first row is data, not header (format: Whatsapp, Nama, NIK)
        dataRows = rows.map((row: any[]) => {
          const whatsapp = String(row[0] || "").trim();
          const name = String(row[1] || "").trim();
          const nik = String(row[2] || "").trim();
          const client = row[3] ? String(row[3] || "").trim() : "";

          return {
            whatsapp: whatsapp || null,
            name: name || null,
            nik: nik || null,
            client: client || null,
          };
        });
      } else {
        // Process data rows (skip header)
        dataRows = rows.slice(1).map((row: any[]) => {
          const whatsapp =
            whatsappIdx >= 0 ? String(row[whatsappIdx] || "").trim() : "";
          const name = nameIdx >= 0 ? String(row[nameIdx] || "").trim() : "";
          const nik = nikIdx >= 0 ? String(row[nikIdx] || "").trim() : "";
          const client = clientIdx >= 0 ? String(row[clientIdx] || "").trim() : "";

          return {
            whatsapp: whatsapp || null,
            name: name || null,
            nik: nik || null,
            client: client || null,
          };
        });
      }

      if (dataRows.length === 0) {
        setError("Tidak ditemukan data.");
        return;
      }

      const res = await fetch("/api/data/bulk", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ data: dataRows }),
      });

      const result = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(result.error || "Gagal import data.");
      }

      setMessage(
        `Berhasil import ${result.inserted} dari ${result.totalSent} data.`,
      );
      
      // Reset file input
      if (input) {
        input.value = "";
        setFileName(null);
      }
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
        Bulk import Data dari Excel
      </h3>
      <p className="mb-4 text-sm text-neutral-500 dark:text-neutral-300">
        Gunakan file Excel (.xlsx, .xls). Format: kolom <strong>Whatsapp</strong>,{" "}
        <strong>Nama</strong>, <strong>NIK</strong> (opsional: <strong>Client</strong>). 
        Client akan dibuat otomatis jika belum ada. Header opsional - jika tidak ada header, 
        akan diasumsikan kolom pertama adalah Whatsapp, kedua Nama, ketiga NIK.
      </p>

      <div className="mb-4">
        <TemplateDownloadButton type="data" />
      </div>

      <form onSubmit={handleSubmit} className="flex flex-wrap items-end gap-3">
        <div>
          <input
            type="file"
            name="file"
            accept=".csv,.xlsx,.xls"
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
