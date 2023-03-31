import { TypeOrmModuleOptions } from '@nestjs/typeorm';

import { DataSourceOptions } from 'typeorm';
import { SnakeNamingStrategy } from 'typeorm-naming-strategies';

export const defaultDataSource: DataSourceOptions = {
  name: 'default',
  type: 'sqlite',
  database: process.env.DATABASE_PATH,
  migrations: [`${__dirname}/migrations/default/**/*{.ts,.js}`],
  synchronize: false,
  migrationsRun: false,
  logging: [undefined, 'development'].includes(process.env.NODE_ENV),
  logger: 'advanced-console',
  namingStrategy: new SnakeNamingStrategy(),
};

export const defaultConfig: TypeOrmModuleOptions = {
  ...defaultDataSource,
  keepConnectionAlive: true,
  autoLoadEntities: true,
};
