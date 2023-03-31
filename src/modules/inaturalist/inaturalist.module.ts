import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';
import inaturalistConfig from './inaturalist.config';
import { INaturalistService } from './inaturalist.service';
import { SeenObservation } from './seen-observation.entity';

@Module({
  imports: [
    ConfigModule.forFeature(inaturalistConfig),
    TypeOrmModule.forFeature([SeenObservation]),
  ],
  providers: [INaturalistService],
  exports: [INaturalistService],
})
export class INaturalistModule {}
