"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";

interface ResetPasswordButtonProps {
  userId: number;
  userEmail: string;
}

export function ResetPasswordButton({
  userId,
  userEmail,
}: ResetPasswordButtonProps) {
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [showModal, setShowModal] = useState(false);
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleReset = async () => {
    if (!password.trim()) {
      setError("Password wajib diisi");
      return;
    }

    if (password.length < 6) {
      setError("Password minimal 6 karakter");
      return;
    }

    setLoading(true);
    setError(null);
    setSuccess(null);

    try {
      const res = await fetch(`/api/users/${userId}/reset-password`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ password: password.trim() }),
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(data.error || "Gagal reset password");
      }

      setSuccess("Password berhasil direset!");
      setPassword("");
      setTimeout(() => {
        setShowModal(false);
        setSuccess(null);
        router.refresh();
      }, 2000);
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan");
    } finally {
      setLoading(false);
    }
  };

  return (
    <>
      <button
        onClick={() => setShowModal(true)}
        className="rounded-md bg-yellow-500 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-yellow-600"
        title="Reset Password"
      >
        Reset Password
      </button>

      {showModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
          <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-lg dark:bg-gray-dark">
            <h3 className="mb-4 text-lg font-semibold text-dark dark:text-white">
              Reset Password
            </h3>
            <p className="mb-4 text-sm text-neutral-600 dark:text-neutral-300">
              Reset password untuk admin: <strong>{userEmail}</strong>
            </p>

            <div className="mb-4">
              <label
                htmlFor="reset_password"
                className="mb-2 block text-sm font-medium text-neutral-600 dark:text-neutral-300"
              >
                Password Baru
              </label>
              <input
                id="reset_password"
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                placeholder="Minimal 6 karakter"
                className="w-full rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
                minLength={6}
                autoFocus
              />
            </div>

            {error && (
              <p className="mb-4 text-sm text-red-500 dark:text-red-400">
                {error}
              </p>
            )}

            {success && (
              <p className="mb-4 text-sm text-emerald-600 dark:text-emerald-400">
                {success}
              </p>
            )}

            <div className="flex gap-3">
              <button
                onClick={() => {
                  setShowModal(false);
                  setPassword("");
                  setError(null);
                  setSuccess(null);
                }}
                className="flex-1 rounded-md border border-stroke px-4 py-2 text-sm font-medium text-neutral-600 transition hover:bg-neutral-50 dark:border-dark-3 dark:text-neutral-300 dark:hover:bg-dark-2"
                disabled={loading}
              >
                Batal
              </button>
              <button
                onClick={handleReset}
                disabled={loading}
                className="flex-1 rounded-md bg-primary px-4 py-2 text-sm font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-60"
              >
                {loading ? "Menyimpan..." : "Reset Password"}
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}
