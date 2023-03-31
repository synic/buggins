import { DataSource } from 'typeorm';

import { defaultDataSource } from './default.config';

const entitiesLocation =
  process.env.NODE_ENV === 'development'
    ? ['src/**/*.entity.ts']
    : ['/app/src/**/*.entity.js'];

export const AppDataSource = new DataSource({
  ...defaultDataSource,
  entities: entitiesLocation,
});
