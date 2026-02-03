import { formatIDR } from "@/lib/currency";
import { compactFormat } from "@/lib/format-number";
import { getOverviewData } from "../../fetch";
import { OverviewCard } from "./card";
import * as icons from "./icons";

type Props = {
  startDate?: Date;
  endDate?: Date;
  clientId?: number;
  isOrganic?: boolean;
};

export async function OverviewCardsGroup(filter?: Props) {
  const { deposit, profit, registrations, clients } = await getOverviewData(filter);

  return (
    <div className="grid gap-4 sm:grid-cols-2 sm:gap-6 xl:grid-cols-4 2xl:gap-7.5">
      <OverviewCard
        label="Total Deposit"
        data={{
          ...deposit,
          value: formatIDR(deposit.value),
        }}
        Icon={icons.Views}
      />

      <OverviewCard
        label="Total Profit"
        data={{
          ...profit,
          value: formatIDR(profit.value),
        }}
        Icon={icons.Profit}
      />

      <OverviewCard
        label="Registrations"
        data={{
          ...registrations,
          value: compactFormat(registrations.value),
        }}
        Icon={icons.Product}
      />

      <OverviewCard
        label="Contacts"
        data={{
          ...clients,
          value: compactFormat(clients.value),
        }}
        Icon={icons.Users}
      />
    </div>
  );
}
