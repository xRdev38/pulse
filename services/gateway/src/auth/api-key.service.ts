import { Injectable } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { createHash } from 'crypto';
import { Pool } from 'pg';

export interface ApiKeyRecord {
  id: string;
  tenantId: string;
  scopes: string[];
}

@Injectable()
export class ApiKeyService {
  private pool: Pool;

  constructor(private config: ConfigService) {
    this.pool = new Pool({
      connectionString: this.config.get<string>('database.url'),
      max: 5,
    });
  }

  async findByHash(rawKey: string): Promise<ApiKeyRecord | null> {
    const hash = createHash('sha256').update(rawKey).digest('hex');
    const { rows } = await this.pool.query(
      'SELECT id, tenant_id, scopes FROM api_keys WHERE key_hash = $1',
      [hash],
    );
    if (!rows.length) return null;
    return { id: rows[0].id, tenantId: rows[0].tenant_id, scopes: rows[0].scopes };
  }

  updateLastUsed(id: string) {
    this.pool
      .query('UPDATE api_keys SET last_used = NOW() WHERE id = $1', [id])
      .catch(() => {});
  }
}
