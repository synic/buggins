import { Module } from '@nestjs/common';
import { TypeOrmModule } from '@nestjs/typeorm';

import { defaultConfig } from './default.config';

@Module({
  imports: [TypeOrmModule.forRoot(defaultConfig)],
})
export class DatabasesModule {}
