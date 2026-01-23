"use client";
import { UserIcon, PasswordIcon } from "@/assets/icons";
import { useRouter, useSearchParams } from "next/navigation";
import Link from "next/link";
import React, { useState } from "react";
import InputGroup from "../FormElements/InputGroup";
import { Checkbox } from "../FormElements/checkbox";
import { signIn } from "next-auth/react";

export default function SigninWithPassword() {
  const router = useRouter();
  const searchParams = useSearchParams();
  const [data, setData] = useState({
    username: "",
    password: "",
    remember: false,
  });

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

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

    try {
      const callbackUrl = searchParams.get("callbackUrl") || "/";
      const result = await signIn("credentials", {
        username: data.username,
        password: data.password,
        redirect: false,
      });

      if (result?.error) {
        // Handle specific error messages
        if (result.error === "CredentialsSignin") {
          setError("Username atau password salah");
        } else {
          setError(result.error || "Terjadi kesalahan saat login");
        }
      } else if (result?.ok) {
        // Login berhasil, redirect ke callback URL atau home
        router.push(callbackUrl);
        router.refresh();
      } else {
        // Fallback jika result tidak jelas
        setError("Terjadi kesalahan saat login");
      }
    } catch (err: any) {
      console.error("Login error:", err);
      setError(err?.message || "Terjadi kesalahan saat login. Pastikan server berjalan dan database terhubung.");
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
        placeholder="Enter your username"
        name="username"
        handleChange={handleChange}
        value={data.username}
        icon={<UserIcon />}
      />

      <InputGroup
        type="password"
        label="Password"
        className="mb-5 [&_input]:py-[15px]"
        placeholder="Enter your password"
        name="password"
        handleChange={handleChange}
        value={data.password}
        icon={<PasswordIcon />}
      />

      <div className="mb-6 flex items-center justify-between gap-2 py-2 font-medium">
        <Checkbox
          label="Remember me"
          name="remember"
          withIcon="check"
          minimal
          radius="md"
          onChange={(e) =>
            setData({
              ...data,
              remember: e.target.checked,
            })
          }
        />

        <Link
          href="/auth/forgot-password"
          className="hover:text-primary dark:text-white dark:hover:text-primary"
        >
          Forgot Password?
        </Link>
      </div>

      <div className="mb-4.5">
        <button
          type="submit"
          className="flex w-full cursor-pointer items-center justify-center gap-2 rounded-lg bg-primary p-4 font-medium text-white transition hover:bg-opacity-90"
        >
          Sign In
          {loading && (
            <span className="inline-block h-4 w-4 animate-spin rounded-full border-2 border-solid border-white border-t-transparent dark:border-primary dark:border-t-transparent" />
          )}
        </button>
      </div>

      {error && (
        <p className="mt-2 text-center text-sm text-red-500 dark:text-red-400">
          {error}
        </p>
      )}
    </form>
  );
}
