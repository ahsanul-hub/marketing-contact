"use client";

import { useRouter } from "next/navigation";
import { useState } from "react";

export function CreateUserForm() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [role, setRole] = useState("user");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    if (!username.trim() || !password.trim()) {
      setError("Username dan password wajib diisi");
      return;
    }

    if (username.length > 30) {
      setError("Username maksimal 30 karakter");
      return;
    }

    if (password.length < 6) {
      setError("Password minimal 6 karakter");
      return;
    }

    setLoading(true);
    try {
      const res = await fetch("/api/users", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: username.trim(),
          password: password.trim(),
          role: role,
        }),
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(data.error || "Gagal membuat admin");
      }

      setUsername("");
      setPassword("");
      setRole("user");
      setSuccess("User berhasil dibuat!");
      router.refresh();

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
          htmlFor="admin_username"
          className="text-neutral-600 dark:text-neutral-300"
        >
          Username
        </label>
        <input
          id="admin_username"
          type="text"
          value={username}
          onChange={(e) => setUsername(e.target.value)}
          placeholder="username"
          maxLength={30}
          className="w-64 rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
          required
        />
      </div>

      <div className="flex flex-col gap-1">
        <label
          htmlFor="admin_password"
          className="text-neutral-600 dark:text-neutral-300"
        >
          Password
        </label>
        <input
          id="admin_password"
          type="password"
          value={password}
          onChange={(e) => setPassword(e.target.value)}
          placeholder="Minimal 6 karakter"
          className="w-48 rounded-md border border-stroke px-3 py-2 text-sm dark:border-dark-3 dark:bg-dark-2"
          required
          minLength={6}
        />
      </div>

      <div className="flex flex-col gap-1">
        <label
          htmlFor="user_role"
          className="text-neutral-600 dark:text-neutral-300"
        >
          Role
        </label>
        <select
          id="user_role"
          value={role}
          onChange={(e) => setRole(e.target.value)}
          className="h-10 w-32 rounded-md border border-stroke px-3 text-sm dark:border-dark-3 dark:bg-dark-2"
        >
          <option value="client">Client</option>
          <option value="user">User</option>
          <option value="admin">Admin</option>
        </select>
      </div>

      <button
        type="submit"
        disabled={loading}
        className="h-10 rounded-md bg-primary px-4 font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-60"
      >
        {loading ? "Menyimpan..." : "Tambah User"}
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
