import { DiscordModule } from '@buggins/discord/discord.module';
import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';
import inaturalistConfig from './inaturalist.config';
import { INaturalistListener } from './inaturalist.listener';
import { INaturalistService } from './inaturalist.service';
import { LoadInatCommand } from './loadinat.command';
import { SeenObservation } from './seen-observation.entity';

@Module({
  imports: [
    ConfigModule.forFeature(inaturalistConfig),
    TypeOrmModule.forFeature([SeenObservation]),
    DiscordModule,
  ],
  providers: [INaturalistService, INaturalistListener, LoadInatCommand],
  exports: [INaturalistService],
})
export class INaturalistModule {}
