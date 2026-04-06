import { Injectable, BadGatewayException, Logger } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import axios, { AxiosInstance } from 'axios';
import { IngestMetricDto } from './ingest.dto';

@Injectable()
export class MetricsService {
  private readonly logger = new Logger(MetricsService.name);
  private readonly http: AxiosInstance;

  constructor(private config: ConfigService) {
    this.http = axios.create({
      baseURL: this.config.get<string>('collectorUrl'),
      timeout: 5000,
      headers: { 'Content-Type': 'application/json' },
    });
  }

  async forward(dto: IngestMetricDto, tenantId: string) {
    // Retry x3 avec backoff exponentiel (Ch.4 Availability)
    for (let attempt = 1; attempt <= 3; attempt++) {
      try {
        const { data } = await this.http.post('/ingest', dto, {
          headers: { 'X-Tenant-ID': tenantId },
        });
        return data;
      } catch (err: any) {
        const retryable = !err.response || err.response.status >= 500;
        if (!retryable || attempt === 3) {
          throw new BadGatewayException('Collector unreachable');
        }
        const delay = 200 * Math.pow(2, attempt - 1);
        this.logger.warn(`Collector retry ${attempt}/3 in ${delay}ms`);
        await new Promise(r => setTimeout(r, delay));
      }
    }
  }
}
