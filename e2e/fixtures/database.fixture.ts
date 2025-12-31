import { test as base, expect } from '@playwright/test';
import { Pool } from 'pg';

const pool = new Pool({
  host: 'localhost',
  port: 5433,
  database: 'agile_party_e2e',
  user: 'postgres',
  password: 'postgres',
});

async function cleanDatabase() {
  const client = await pool.connect();
  try {
    await client.query('DELETE FROM tasks');
    await client.query('DELETE FROM rooms');
  } finally {
    client.release();
  }
}

type DatabaseFixtures = {
  cleanDatabase: void;
};

export const test = base.extend<DatabaseFixtures>({
  cleanDatabase: [async ({}, use) => {
    await cleanDatabase();
    await use();
    await cleanDatabase();
  }, { auto: true }],
});

export { expect };
