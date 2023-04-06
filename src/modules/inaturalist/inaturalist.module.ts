import { DiscordModule } from '@buggins/discord/discord.module';
import { Module } from '@nestjs/common';
import { ConfigModule } from '@nestjs/config';
import { TypeOrmModule } from '@nestjs/typeorm';
import { INaturalistSettingsEntity } from './inaturalist-settings.entity';
import inaturalistConfig from './inaturalist.config';
import { INaturalistService } from './inaturalist.service';
import { LoadInatCommand } from './loadinat.command';
import { SeenObservationEntity } from './seen-observation.entity';

@Module({
  imports: [
    ConfigModule.forFeature(inaturalistConfig),
    TypeOrmModule.forFeature([
      INaturalistSettingsEntity,
      SeenObservationEntity,
    ]),
    DiscordModule,
  ],
  providers: [INaturalistService, LoadInatCommand],
  exports: [INaturalistService],
})
export class INaturalistModule {}
