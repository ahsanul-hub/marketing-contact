/**
 * Pagination & Date Range Utilities
 * 
 * Utility functions untuk parsing pagination dan date range dari URL search params.
 * Digunakan di halaman-halaman yang memiliki tabel dengan pagination.
 * 
 * Features:
 * - Parse page dan limit dengan validation
 * - Parse start date dan end date
 * - Default values untuk pagination
 */

export const MIN_LIMIT = 10;
export const MAX_LIMIT = 10000;

type RawParam = string | string[] | undefined;

type PaginationSearchParams = {
  page?: RawParam;
  limit?: RawParam;
  start?: RawParam;
  end?: RawParam;
};

/**
 * Parse pagination parameters dari URL search params
 * @param searchParams - URL search params object
 * @returns Object dengan page dan limit yang sudah di-validate
 */
export function parsePaginationParams(
  searchParams?: PaginationSearchParams,
): { page: number; limit: number } {
  const rawPage = normalizeParam(searchParams?.page);
  const rawLimit = normalizeParam(searchParams?.limit);

  const pageFromQuery = rawPage ? Number(rawPage) : 1;
  const limitFromQuery = rawLimit ? Number(rawLimit) : MIN_LIMIT;

  const page = Number.isFinite(pageFromQuery) ? Math.max(1, pageFromQuery) : 1;
  const unclampedLimit = Number.isFinite(limitFromQuery)
    ? limitFromQuery
    : MIN_LIMIT;
  const limit = clamp(unclampedLimit, MIN_LIMIT, MAX_LIMIT);

  return { page, limit };
}

/**
 * Parse date range parameters dari URL search params
 * @param searchParams - URL search params object
 * @returns Object dengan startDate, endDate, startParam, endParam
 * 
 * Note: startParam dan endParam digunakan untuk defaultValue di input date
 */
export function parseDateRangeParams(
  searchParams?: PaginationSearchParams,
): {
  startDate?: Date;
  endDate?: Date;
  startParam?: string;
  endParam?: string;
} {
  const rawStart =
    normalizeParam(searchParams?.start) ??
    normalizeParam((searchParams as any)?.start_date);
  const rawEnd =
    normalizeParam(searchParams?.end) ??
    normalizeParam((searchParams as any)?.end_date);

  const startDate = rawStart && isValidDate(rawStart) ? new Date(rawStart) : undefined;
  const endDate = rawEnd && isValidDate(rawEnd) ? new Date(rawEnd) : undefined;

  return {
    startDate,
    endDate,
    startParam: rawStart ?? undefined,
    endParam: rawEnd ?? undefined,
  };
}

function normalizeParam(param: RawParam) {
  if (!param) return undefined;
  return Array.isArray(param) ? param[0] : param;
}

function clamp(value: number, min: number, max: number) {
  return Math.min(max, Math.max(min, value));
}

function isValidDate(value: string) {
  const date = new Date(value);
  return !Number.isNaN(date.getTime());
}
