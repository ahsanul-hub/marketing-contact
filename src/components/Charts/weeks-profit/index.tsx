import { PeriodPicker } from "@/components/period-picker";
import { cn } from "@/lib/utils";
import { AnalyticsFilter, getWeeklyProfitBars } from "@/services/analytics";
import { WeeksProfitChart } from "./chart";

type PropsType = {
  timeFrame?: string;
  className?: string;
  filter?: AnalyticsFilter;
};

export async function WeeksProfit({ className, timeFrame, filter }: PropsType) {
  const data = await getWeeklyProfitBars(timeFrame, filter);

  return (
    <div
      className={cn(
        "rounded-[10px] bg-white px-7.5 pt-7.5 shadow-1 dark:bg-gray-dark dark:shadow-card",
        className,
      )}
    >
      <div className="flex flex-wrap items-center justify-between gap-4">
        <h2 className="text-body-2xlg font-bold text-dark dark:text-white">
          Deposit vs Profit {timeFrame || "this week"}
        </h2>

        <PeriodPicker
          items={["this week", "last week"]}
          defaultValue={timeFrame || "this week"}
          sectionKey="weeks_profit"
        />
      </div>

      <WeeksProfitChart data={data} />
    </div>
  );
}
