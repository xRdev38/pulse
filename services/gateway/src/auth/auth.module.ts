import { Module } from '@nestjs/common';
import { ApiKeyGuard } from './api-key.guard';
import { ApiKeyService } from './api-key.service';

@Module({
  providers: [ApiKeyGuard, ApiKeyService],
  exports: [ApiKeyGuard, ApiKeyService],
})
export class AuthModule {}
