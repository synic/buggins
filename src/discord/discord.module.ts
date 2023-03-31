import { ConfigModule } from '@nestjs/config';
import { Module } from '@nestjs/common';
import discordConfig from './discord.config';
import { DiscordService } from './discord.service';

@Module({
  imports: [ConfigModule.forFeature(discordConfig)],
  providers: [DiscordService],
  exports: [DiscordService],
})
export class DiscordModule {}
