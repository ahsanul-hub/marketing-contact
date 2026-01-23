"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export function CreateClientForm() {
  const router = useRouter();
  const [name, setName] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);
    const trimmed = name.trim();
    if (!trimmed) {
      setError("Nama client wajib diisi");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch("/api/clients", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ name: trimmed }),
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(data.error || "Gagal membuat client");
      }

      setName("");
      setSuccess("Client berhasil dibuat!");
      // Refresh data on the page
      router.refresh();
      
      // Clear success message after 3 seconds
      setTimeout(() => setSuccess(null), 3000);
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form
      onSubmit={handleSubmit}
      className="mb-4 flex flex-wrap items-end gap-3 rounded-[10px] border border-stroke bg-white p-4 text-sm shadow-1 dark:border-dark-3 dark:bg-gray-dark dark:shadow-card"
    >
      <div className="flex flex-col gap-1">
        <label
          htmlFor="client_name"
          className="text-neutral-600 dark:text-neutral-300"
        >
          Tambah Client
        </label>
        <input
          id="client_name"
          name="client_name"
          type="text"
          value={name}
          onChange={(e) => setName(e.target.value)}
          placeholder="Nama client"
          className="w-64 rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
        />
      </div>

      <button
        type="submit"
        disabled={loading}
        className="h-10 rounded-md bg-primary px-4 font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-60"
      >
        {loading ? "Menyimpan..." : "Simpan"}
      </button>

      {error && (
        <p className="mt-2 w-full text-sm text-red-500 dark:text-red-400">
          {error}
        </p>
      )}

      {success && (
        <p className="mt-2 w-full text-sm text-emerald-600 dark:text-emerald-400">
          {success}
        </p>
      )}
    </form>
  );
}

