/*
  Warnings:

  - A unique constraint covering the columns `[name]` on the table `client` will be added. If there are existing duplicate values, this will fail.
  - A unique constraint covering the columns `[id_client,phone_number]` on the table `registration` will be added. If there are existing duplicate values, this will fail.
  - Made the column `name` on table `client` required. This step will fail if there are existing NULL values in that column.
  - Added the required column `id_client` to the `registration` table without a default value. This is not possible if the table is not empty.
  - Added the required column `id_client` to the `transaction` table without a default value. This is not possible if the table is not empty.

*/
-- AlterTable
ALTER TABLE "client" ALTER COLUMN "name" SET NOT NULL;

-- AlterTable
ALTER TABLE "registration" ADD COLUMN     "id_client" BIGINT NOT NULL;

-- AlterTable
ALTER TABLE "transaction" ADD COLUMN     "id_client" BIGINT NOT NULL;

-- CreateIndex
CREATE UNIQUE INDEX "client_name_key" ON "client"("name");

-- CreateIndex
CREATE UNIQUE INDEX "registration_id_client_phone_number_key" ON "registration"("id_client", "phone_number");

-- AddForeignKey
ALTER TABLE "registration" ADD CONSTRAINT "registration_id_client_fkey" FOREIGN KEY ("id_client") REFERENCES "client"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "transaction" ADD CONSTRAINT "transaction_id_client_fkey" FOREIGN KEY ("id_client") REFERENCES "client"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
