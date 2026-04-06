import 'reflect-metadata';
import { NestFactory } from '@nestjs/core';
import { ValidationPipe } from '@nestjs/common';
import { AppModule } from './app.module';

async function bootstrap() {
  const app = await NestFactory.create(AppModule);

  // Préfixe global : toutes les routes → /api/...
  app.setGlobalPrefix('api');

  // Validation automatique des DTOs (class-validator)
  app.useGlobalPipes(
    new ValidationPipe({ whitelist: true, transform: true }),
  );

  // /health est hors du préfixe /api (accessible directement)
  // NestJS : les controllers sans préfixe de module héritent quand même du global prefix
  // → on exclut /health du préfixe
  app.setGlobalPrefix('api', { exclude: ['health'] });

  const port = process.env.PORT ?? 3000;
  await app.listen(port);
  console.log(`🚀 Gateway running on :${port}`);
}
bootstrap();
