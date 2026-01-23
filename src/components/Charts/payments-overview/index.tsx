import { PeriodPicker } from "@/components/period-picker";
import { formatIDR } from "@/lib/currency";
import { cn } from "@/lib/utils";
import { AnalyticsFilter, getDepositProfitSeries } from "@/services/analytics";
import { PaymentsOverviewChart } from "./chart";

type PropsType = {
  timeFrame?: string;
  className?: string;
  filter?: AnalyticsFilter;
};

export async function PaymentsOverview({
  timeFrame = "monthly",
  className,
  filter,
}: PropsType) {
  const data = await getDepositProfitSeries(timeFrame, filter);

  return (
    <div
      className={cn(
        "grid gap-2 rounded-[10px] bg-white px-7.5 pb-6 pt-7.5 shadow-1 dark:bg-gray-dark dark:shadow-card",
        className,
      )}
    >
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-body-2xlg font-bold text-dark dark:text-white">
          Deposit & Profit Trend
        </h2>

        <PeriodPicker defaultValue={timeFrame} sectionKey="payments_overview" />
      </div>

      <PaymentsOverviewChart data={data} />

      <dl className="grid divide-stroke text-center dark:divide-dark-3 sm:grid-cols-2 sm:divide-x [&>div]:flex [&>div]:flex-col-reverse [&>div]:gap-1">
        <div className="dark:border-dark-3 max-sm:mb-3 max-sm:border-b max-sm:pb-3">
          <dt className="text-xl font-bold text-dark dark:text-white">
            {formatIDR(data.deposit.reduce((acc, { y }) => acc + y, 0))}
          </dt>
          <dd className="font-medium dark:text-dark-6">Total Deposit</dd>
        </div>

        <div>
          <dt className="text-xl font-bold text-dark dark:text-white">
            {formatIDR(data.profit.reduce((acc, { y }) => acc + y, 0))}
          </dt>
          <dd className="font-medium dark:text-dark-6">Total Profit</dd>
        </div>
      </dl>
    </div>
  );
}
