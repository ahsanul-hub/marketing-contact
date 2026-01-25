"use client";
import { UserIcon, PasswordIcon } from "@/assets/icons";
import { useRouter } from "next/navigation";
import Link from "next/link";
import React, { useState } from "react";
import InputGroup from "../FormElements/InputGroup";
import { signIn } from "next-auth/react";

export default function SignupWithPassword() {
  const router = useRouter();
  const [data, setData] = useState({
    username: "",
    password: "",
    confirmPassword: "",
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState(false);

  const handleChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setData({
      ...data,
      [e.target.name]: e.target.value,
    });
  };

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    setError(null);
    setSuccess(false);

    // Validasi
    if (data.password !== data.confirmPassword) {
      setError("Password dan konfirmasi password tidak sama");
      setLoading(false);
      return;
    }

    try {
      // Register user
      const res = await fetch("/api/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          username: data.username,
          password: data.password,
        }),
      });

      const result = await res.json();

      if (!res.ok) {
        throw new Error(result.error || "Gagal melakukan registrasi");
      }

      // Registrasi berhasil, langsung login
      setSuccess(true);
      
      // Auto login setelah registrasi
      const loginResult = await signIn("credentials", {
        username: data.username,
        password: data.password,
        redirect: false,
      });

      if (loginResult?.error) {
        // Jika auto login gagal, redirect ke sign in
        router.push("/auth/sign-in?registered=true");
      } else if (loginResult?.ok) {
        // Login berhasil, redirect ke home
        router.push("/");
        router.refresh();
      }
    } catch (err: any) {
      console.error("Register error:", err);
      setError(err?.message || "Terjadi kesalahan saat registrasi");
    } finally {
      setLoading(false);
    }
  };

  return (
    <form onSubmit={handleSubmit}>
      <InputGroup
        type="text"
        label="Username"
        className="mb-4 [&_input]:py-[15px]"
        placeholder="Masukkan username"
        name="username"
        handleChange={handleChange}
        value={data.username}
        icon={<UserIcon />}
        required
      />

      <InputGroup
        type="password"
        label="Password"
        className="mb-4 [&_input]:py-[15px]"
        placeholder="Masukkan password"
        name="password"
        handleChange={handleChange}
        value={data.password}
        icon={<PasswordIcon />}
        required
      />

      <InputGroup
        type="password"
        label="Konfirmasi Password"
        className="mb-5 [&_input]:py-[15px]"
        placeholder="Konfirmasi password"
        name="confirmPassword"
        handleChange={handleChange}
        value={data.confirmPassword}
        icon={<PasswordIcon />}
        required
      />

      <div className="mb-4.5">
        <button
          type="submit"
          disabled={loading}
          className="flex w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-primary p-4 font-medium text-white transition hover:bg-opacity-90 disabled:cursor-not-allowed disabled:bg-opacity-60"
        >
          Daftar
          {loading && (
            <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-solid border-white border-t-transparent dark:border-primary dark:border-t-transparent" />
          )}
        </button>
      </div>

      {error && (
        <p className="mb-4 text-center text-sm text-red-500 dark:text-red-400">
          {error}
        </p>
      )}

      {success && (
        <p className="mb-4 text-center text-sm text-emerald-600 dark:text-emerald-400">
          Registrasi berhasil! Mengarahkan ke halaman utama...
        </p>
      )}

      <div className="text-center text-sm">
        <span className="text-neutral-600 dark:text-neutral-300">
          Sudah punya akun?{" "}
        </span>
        <Link
          href="/auth/sign-in"
          className="font-medium text-primary hover:underline"
        >
          Masuk
        </Link>
      </div>
    </form>
  );
}

