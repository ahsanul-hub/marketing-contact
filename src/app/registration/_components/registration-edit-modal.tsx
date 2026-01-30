"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

interface RegistrationEditModalProps {
  isOpen: boolean;
  onClose: () => void;
  registrationId: bigint;
  phoneNumber: string;
  clientName?: string;
  createdAt?: string | null;
}

export function RegistrationEditModal({
  isOpen,
  onClose,
  registrationId,
  phoneNumber,
  clientName,
  createdAt,
}: RegistrationEditModalProps) {
  const router = useRouter();
  const [newPhoneNumber, setNewPhoneNumber] = useState(phoneNumber);
  const formatToLocalDatetime = (iso?: string | null) => {
    if (!iso) return "";
    const d = new Date(iso);
    if (isNaN(d.getTime())) return "";
    // yyyy-mm-ddThh:mm (datetime-local)
    const pad = (n: number) => n.toString().padStart(2, "0");
    const yyyy = d.getFullYear();
    const mm = pad(d.getMonth() + 1);
    const dd = pad(d.getDate());
    const hh = pad(d.getHours());
    const mi = pad(d.getMinutes());
    return `${yyyy}-${mm}-${dd}T${hh}:${mi}`;
  };

  const [newCreatedAt, setNewCreatedAt] = useState<string>(
    formatToLocalDatetime(createdAt)
  );
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsLoading(true);
    setError(null);

    try {
      const response = await fetch("/api/registration/update", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: registrationId.toString(),
          phoneNumber: newPhoneNumber,
          createdAt: newCreatedAt || undefined,
        }),
      });

      if (!response.ok) {
        // try parse json message, fallback to text
        let msg = "Gagal update data";
        try {
          const errorData = await response.json();
          msg = errorData?.message || msg;
        } catch (e) {
          try {
            const txt = await response.text();
            if (txt) msg = txt;
          } catch (e) {
            /* ignore */
          }
        }
        throw new Error(msg);
      }

      onClose();
      router.refresh();
    } catch (err) {
      setError(err instanceof Error ? err.message : "Terjadi kesalahan");
    } finally {
      setIsLoading(false);
    }
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-lg dark:bg-gray-dark">
        <div className="mb-4">
          <h2 className="text-lg font-semibold text-dark dark:text-white">
            Edit Registrasi
          </h2>
          <p className="mt-1 text-sm text-neutral-500 dark:text-neutral-300">
            Update nomor telepon registrasi
          </p>
        </div>

        {error && (
          <div className="mb-4 rounded bg-red-100 p-3 text-sm text-red-700 dark:bg-red-900/20 dark:text-red-200">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <div>
            <label className="mb-2 block text-sm font-medium text-dark dark:text-white">
              Nomor Telepon
            </label>
            <input
              type="text"
              value={newPhoneNumber}
              onChange={(e) => setNewPhoneNumber(e.target.value)}
              className="w-full rounded border border-stroke bg-white px-3 py-2 text-dark outline-none transition dark:border-dark-3 dark:bg-dark-2 dark:text-white"
              placeholder="Masukkan nomor telepon"
              disabled={isLoading}
            />
          </div>

          <div>
            <label className="mb-2 block text-sm font-medium text-dark dark:text-white">
              Tanggal Dibuat
            </label>
            <input
              type="datetime-local"
              value={newCreatedAt}
              onChange={(e) => setNewCreatedAt(e.target.value)}
              className="w-full rounded border border-stroke bg-white px-3 py-2 text-dark outline-none transition dark:border-dark-3 dark:bg-dark-2 dark:text-white"
              disabled={isLoading}
            />
          </div>

          {clientName && (
            <div>
              <label className="mb-2 block text-sm font-medium text-dark dark:text-white">
                Klien
              </label>
              <input
                type="text"
                value={clientName}
                disabled
                className="w-full rounded border border-stroke bg-gray-100 px-3 py-2 text-dark outline-none dark:border-dark-3 dark:bg-dark-3 dark:text-neutral-300"
              />
            </div>
          )}

          <div className="flex gap-3 pt-4">
            <button
              type="button"
              onClick={onClose}
              disabled={isLoading}
              className="flex-1 rounded border border-stroke px-4 py-2 font-medium text-dark transition hover:border-primary hover:text-primary disabled:opacity-50 dark:border-dark-3 dark:text-white"
            >
              Batal
            </button>
            <button
              type="submit"
              disabled={isLoading}
              className="flex-1 rounded bg-primary px-4 py-2 font-medium text-white transition hover:bg-opacity-90 disabled:opacity-50"
            >
              {isLoading ? "Menyimpan..." : "Simpan"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
