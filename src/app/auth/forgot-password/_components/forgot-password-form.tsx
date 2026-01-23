"use client";

import { UserIcon, PasswordIcon } from "@/assets/icons";
import { useRouter } from "next/navigation";
import { useState } from "react";
import InputGroup from "@/components/FormElements/InputGroup";

export default function ForgotPasswordForm() {
  const router = useRouter();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError(null);
    setSuccess(null);

    // Validasi
    if (!username.trim() || !password.trim() || !confirmPassword.trim()) {
      setError("Semua field wajib diisi");
      return;
    }

    if (password.length < 6) {
      setError("Password minimal 6 karakter");
      return;
    }

    if (password !== confirmPassword) {
      setError("Password dan konfirmasi password tidak sama");
      return;
    }

    setLoading(true);

    try {
      const res = await fetch("/api/reset-password", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: username.trim(),
          password: password.trim(),
        }),
      });

      const data = await res.json().catch(() => ({}));

      if (!res.ok) {
        throw new Error(data.error || "Gagal reset password");
      }

      setSuccess("Password berhasil direset! Redirecting...");
      
      // Redirect ke login setelah 2 detik
      setTimeout(() => {
        router.push("/auth/sign-in");
      }, 2000);
    } catch (err: any) {
      setError(err.message || "Terjadi kesalahan saat reset password");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit} className="mt-8 space-y-6">
      <div className="space-y-4">
        <InputGroup
          type="text"
          label="Username"
          className="[&_input]:py-[15px]"
          placeholder="Enter your username"
          name="username"
          handleChange={(e) => setUsername(e.target.value)}
          value={username}
          icon={<UserIcon />}
          required
        />

        <InputGroup
          type="password"
          label="Password Baru"
          className="[&_input]:py-[15px]"
          placeholder="Minimal 6 karakter"
          name="password"
          handleChange={(e) => setPassword(e.target.value)}
          value={password}
          icon={<PasswordIcon />}
          required
          minLength={6}
        />

        <InputGroup
          type="password"
          label="Konfirmasi Password"
          className="[&_input]:py-[15px]"
          placeholder="Ulangi password baru"
          name="confirmPassword"
          handleChange={(e) => setConfirmPassword(e.target.value)}
          value={confirmPassword}
          icon={<PasswordIcon />}
          required
          minLength={6}
        />
      </div>

      {error && (
        <div className="rounded-md bg-red-50 p-4 dark:bg-red-900/20">
          <p className="text-sm text-red-600 dark:text-red-400">{error}</p>
        </div>
      )}

      {success && (
        <div className="rounded-md bg-emerald-50 p-4 dark:bg-emerald-900/20">
          <p className="text-sm text-emerald-600 dark:text-emerald-400">
            {success}
          </p>
        </div>
      )}

      <div>
        <button
          type="submit"
          disabled={loading}
          className="flex w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-primary p-4 font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-60"
        >
          {loading ? (
            <>
              <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-solid border-white border-t-transparent" />
              Memproses...
            </>
          ) : (
            "Reset Password"
          )}
        </button>
      </div>
    </form>
  );
}
