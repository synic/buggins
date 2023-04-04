import { LogLevel } from '@nestjs/common';
import { NestExpressApplication } from '@nestjs/platform-express';
import { NestFactory } from '@nestjs/core';
import { AppModule } from './app.module';

async function bootstrap() {
  const logger: LogLevel[] = ['verbose', 'log', 'warn', 'error'];
  if (process.env.NODE_ENV === 'development') logger.push('debug');

  const app = await NestFactory.create<NestExpressApplication>(AppModule, {
    logger,
  });
  await app.listen(3000);
}
bootstrap();
