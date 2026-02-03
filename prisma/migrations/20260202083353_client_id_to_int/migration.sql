/*
  Warnings:

  - The primary key for the `client` table will be changed. If it partially fails, the table could be left without primary key constraint.
  - You are about to alter the column `id` on the `client` table. The data in that column could be lost. The data in that column will be cast from `BigInt` to `Integer`.
  - You are about to drop the column `id_client` on the `data` table. All the data in the column will be lost.
  - You are about to alter the column `id_client` on the `registration` table. The data in that column could be lost. The data in that column will be cast from `BigInt` to `Integer`.
  - You are about to alter the column `id_client` on the `transaction` table. The data in that column could be lost. The data in that column will be cast from `BigInt` to `Integer`.

*/
-- DropForeignKey
ALTER TABLE "data" DROP CONSTRAINT "data_id_client_fkey";

-- DropForeignKey
ALTER TABLE "registration" DROP CONSTRAINT "registration_id_client_fkey";

-- DropForeignKey
ALTER TABLE "transaction" DROP CONSTRAINT "transaction_id_client_fkey";

-- AlterTable
ALTER TABLE "client"
DROP CONSTRAINT "client_pkey",
ALTER COLUMN "id" TYPE INTEGER USING "id"::INTEGER,
ADD CONSTRAINT "client_pkey" PRIMARY KEY ("id");

-- AlterTable
ALTER TABLE "data" DROP COLUMN "id_client",
ADD COLUMN     "owner_name" TEXT;

-- AlterTable
ALTER TABLE "registration" ALTER COLUMN "id_client" SET DATA TYPE INTEGER;

-- AlterTable
ALTER TABLE "transaction" ALTER COLUMN "id_client" SET DATA TYPE INTEGER;

-- AddForeignKey
ALTER TABLE "registration" ADD CONSTRAINT "registration_id_client_fkey" FOREIGN KEY ("id_client") REFERENCES "client"("id") ON DELETE RESTRICT ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "transaction" ADD CONSTRAINT "transaction_id_client_fkey" FOREIGN KEY ("id_client") REFERENCES "client"("id") ON DELETE RESTRICT ON UPDATE CASCADE;
