import {
  Controller, Post, Body, Req,
  UseGuards, HttpCode, HttpStatus,
} from '@nestjs/common';
import { ApiKeyGuard } from '../auth/api-key.guard';
import { MetricsService } from './metrics.service';
import { IngestMetricDto } from './ingest.dto';

@Controller('metrics')
@UseGuards(ApiKeyGuard)
export class MetricsController {
  constructor(private readonly svc: MetricsService) {}

  @Post('ingest')
  @HttpCode(HttpStatus.ACCEPTED)
  ingest(@Body() dto: IngestMetricDto, @Req() req: any) {
    return this.svc.forward(dto, req.tenantId);
  }
}
