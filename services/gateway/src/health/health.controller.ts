import { Controller, Get } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import axios from 'axios';

@Controller()
export class HealthController {
  constructor(private config: ConfigService) {}

  @Get('health')
  async check() {
    const checks: Record<string, any> = {};

    try {
      const t = Date.now();
      await axios.get(`${this.config.get('collectorUrl')}/health`, { timeout: 2000 });
      checks['collector'] = { status: 'ok', latency: `${Date.now() - t}ms` };
    } catch (e: any) {
      checks['collector'] = { status: 'error', error: e.message };
    }

    const allOk = Object.values(checks).every((c: any) => c.status === 'ok');
    return {
      status: allOk ? 'healthy' : 'degraded',
      checks,
      uptime: `${Math.floor(process.uptime())}s`,
    };
  }
}
