import { TypeOrmModuleOptions } from '@nestjs/typeorm';

import { DataSourceOptions } from 'typeorm';
import { SnakeNamingStrategy } from 'typeorm-naming-strategies';

type DatabaseType = 'postgres' | 'sqlite' | 'mysql';
const type: DatabaseType = (process.env.DATABASE_TYPE ??
  'sqlite') as DatabaseType;

const urlConfig =
  type === 'sqlite'
    ? { database: process.env.DATABASE_URL ?? '' }
    : { url: process.env.DATABASE_URL ?? '' };

export const defaultDataSource: DataSourceOptions = {
  name: 'default',
  type,
  database: '',
  ...urlConfig,
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
