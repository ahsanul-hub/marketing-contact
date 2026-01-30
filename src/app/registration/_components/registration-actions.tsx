"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import { RegistrationEditModal } from "./registration-edit-modal";

interface RegistrationActionsProps {
  registrationId: bigint;
  phoneNumber: string;
  clientName?: string;
  createdAt?: string | null;
}

export function RegistrationActions({
  registrationId,
  phoneNumber,
  clientName,
  createdAt,
}: RegistrationActionsProps) {
  const router = useRouter();
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [isDeleting, setIsDeleting] = useState(false);

  const handleDelete = async () => {
    if (!confirm("Apakah Anda yakin ingin menghapus data ini?")) {
      return;
    }

    setIsDeleting(true);

    try {
      const response = await fetch("/api/registration/delete", {
        method: "DELETE",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          id: registrationId.toString(),
        }),
      });

      if (!response.ok) {
        const errorData = await response.json();
        alert(`Error: ${errorData.message || "Gagal menghapus data"}`);
        return;
      }

      router.refresh();
    } catch (error) {
      alert(
        error instanceof Error ? error.message : "Terjadi kesalahan saat menghapus"
      );
    } finally {
      setIsDeleting(false);
    }
  };

  return (
    <>
      <div className="flex items-center gap-2">
        <button
          onClick={() => setIsEditModalOpen(true)}
          className="inline-flex h-8 w-8 items-center justify-center rounded bg-blue-100 text-blue-600 hover:bg-blue-200 dark:bg-blue-900/30 dark:text-blue-300"
          title="Edit"
        >
          <svg
            className="h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
            />
          </svg>
        </button>
        <button
          onClick={handleDelete}
          disabled={isDeleting}
          className="inline-flex h-8 w-8 items-center justify-center rounded bg-red-100 text-red-600 hover:bg-red-200 disabled:opacity-50 dark:bg-red-900/30 dark:text-red-300"
          title="Delete"
        >
          <svg
            className="h-4 w-4"
            fill="none"
            stroke="currentColor"
            viewBox="0 0 24 24"
          >
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={2}
              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
            />
          </svg>
        </button>
      </div>

      <RegistrationEditModal
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        registrationId={registrationId}
        phoneNumber={phoneNumber}
        clientName={clientName}
        createdAt={createdAt}
      />
    </>
  );
}
