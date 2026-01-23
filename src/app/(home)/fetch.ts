/**
 * Fetch functions for home page data
 * 
 * This file contains data fetching functions used by the home/dashboard page.
 * All data is fetched from the analytics service which queries the database.
 */

import { AnalyticsFilter, getOverviewMetrics } from "@/services/analytics";

/**
 * Get overview metrics data (deposit, profit, registrations, contacts)
 * @param filter - Optional filter for date range and client filtering
 * @returns Overview metrics data
 */
export async function getOverviewData(filter?: AnalyticsFilter) {
  return getOverviewMetrics(filter);
}