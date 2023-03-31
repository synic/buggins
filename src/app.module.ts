import { Module } from '@nestjs/common';
import { DatabasesModule } from './databases/databases.module';
import { DiscordModule } from './discord/discord.module';
import { INaturalistModule } from './modules/inaturalist/inaturalist.module';

const enabledModules = [INaturalistModule];

@Module({
  imports: [DatabasesModule, DiscordModule, ...enabledModules],
})
export class AppModule {}
