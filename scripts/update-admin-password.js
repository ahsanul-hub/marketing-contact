/**
 * Script untuk update password admin di database
 *
 * Usage:
 * node scripts/update-admin-password.js fahri newpassword123
 */

const { PrismaClient } = require('@prisma/client');
const bcrypt = require('bcryptjs');

const prisma = new PrismaClient();

async function updateAdminPassword() {
  const username = process.argv[2];
  const password = process.argv[3] || 'password123';

  if (!username) {
    console.error('Usage: node scripts/update-admin-password.js <username> [password]');
    process.exit(1);
  }

  try {
    // Hash password
    const hashedPassword = await bcrypt.hash(password, 10);
    // Convert to Buffer for bytea
    const passwordBuffer = Buffer.from(hashedPassword, 'utf-8');

    console.log('\n=== Update Admin Password ===');
    console.log('Username:', username);
    console.log('Password:', password);
    console.log('\nHashed Password:');
    console.log(hashedPassword);

    // Update password di database
    const admin = await prisma.admin.update({
      where: { username },
      data: {
        password: passwordBuffer,
        updated_at: new Date(),
      },
      select: {
        id: true,
        username: true,
        role: true,
        active: true,
      },
    });

    console.log('\n✅ Admin password updated successfully!');
    console.log('Admin:', admin);
    console.log('\n=== SQL Query (for reference) ===');
    console.log(`UPDATE admin SET password = '\\x${passwordBuffer.toString('hex')}', updated_at = NOW() WHERE username = '${username}';`);
    console.log('\n');
  } catch (error) {
    console.error('Error updating admin password:', error);
    if (error.code === 'P2025') {
      console.error('Admin not found');
    }
  } finally {
    await prisma.$disconnect();
  }
}

updateAdminPassword();