import {
  CanActivate,
  ExecutionContext,
  Injectable,
  UnauthorizedException,
} from '@nestjs/common';
import { ApiKeyService } from './api-key.service';

@Injectable()
export class ApiKeyGuard implements CanActivate {
  constructor(private readonly svc: ApiKeyService) {}

  async canActivate(context: ExecutionContext): Promise<boolean> {
    const req = context.switchToHttp().getRequest();
    const key = req.headers['x-api-key'] as string;
    if (!key) throw new UnauthorizedException('Missing X-API-Key header');

    const record = await this.svc.findByHash(key);
    if (!record) throw new UnauthorizedException('Invalid API key');

    req.tenantId = record.tenantId;
    this.svc.updateLastUsed(record.id);
    return true;
  }
}
